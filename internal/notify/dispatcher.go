package notify

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/uptrace/bun"
)

const (
	tickInterval   = 30 * time.Second
	channelWidth   = 2  // concurrent sends per channel; also the claim batch size
	perChannelTick = 20 // rows one channel may consume per tick, so a flood cannot starve the loop
	leaseSlack     = 20 * time.Second
)

// Schedule supplies the user's timezone and notify hour for date-only reminders.
type Schedule func(ctx context.Context) (*time.Location, int)

// Dispatcher turns due reminders into per-channel deliveries and sends them.
// Delivery is at-least-once: a crash between a successful send and the `sent`
// write can duplicate one notification.
type Dispatcher struct {
	DB          *bun.DB
	Registry    *Registry
	Channels    *ChannelRepo
	Client      *HTTPClient
	Schedule    Schedule
	Now         func() time.Time
	SendTimeout time.Duration
}

func NewDispatcher(db *bun.DB, reg *Registry, chans *ChannelRepo, client *HTTPClient, schedule Schedule) *Dispatcher {
	return &Dispatcher{
		DB: db, Registry: reg, Channels: chans, Client: client, Schedule: schedule,
		Now: time.Now, SendTimeout: attemptTimeout,
	}
}

// Run ticks until ctx is cancelled. The returned channel closes once the
// in-flight tick (including its sends) has finished.
func (d *Dispatcher) Run(ctx context.Context) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)

		ticker := time.NewTicker(tickInterval)
		defer ticker.Stop()

		d.Tick(ctx)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				d.Tick(ctx)
			}
		}
	}()

	return done
}

// Tick runs one synchronous enqueue, cancel, dispatch and cleanup cycle.
func (d *Dispatcher) Tick(ctx context.Context) {
	now := d.Now().UTC().Truncate(time.Second)

	recs, err := d.Channels.List(ctx)
	if err != nil {
		slog.Error("notify: list channels", "error", err)
		return
	}

	if err := d.enqueue(ctx, now, recs); err != nil {
		slog.Error("notify: enqueue", "error", err)
	}

	if err := cancelDead(ctx, d.DB, now); err != nil {
		slog.Error("notify: cancel", "error", err)
	}

	var wg sync.WaitGroup

	for _, rec := range recs {
		if !rec.Enabled {
			continue
		}

		wg.Add(1)

		go func() {
			defer wg.Done()

			d.drainChannel(ctx, rec)
		}()
	}

	wg.Wait()

	if err := housekeeping(context.WithoutCancel(ctx), d.DB, now); err != nil {
		slog.Warn("notify: housekeeping", "error", err)
	}
}

func (d *Dispatcher) enqueue(ctx context.Context, now time.Time, recs []*Record) error {
	var targets []*Record

	for _, rec := range recs {
		if rec.Enabled && !rec.Unreadable {
			targets = append(targets, rec)
		}
	}

	if len(targets) == 0 {
		return nil
	}

	cands, err := candidates(ctx, d.DB, now)
	if err != nil {
		return err
	}

	loc, hour := d.Schedule(ctx)

	var rows []*Delivery

	for _, c := range cands {
		fire := FireTime(c.DueDate, loc, hour).UTC().Truncate(time.Second)

		// Bulk imports stamp "today"; a reminder created after its fire time never notifies.
		// Judged against fire (not the stored 00:00 UTC) so a date-only reminder made earlier
		// on its local due day still fires.
		if !c.CreatedAt.Before(fire) {
			continue
		}

		if !fire.After(now.Add(-staleAge)) || fire.After(now) {
			continue
		}

		for _, rec := range targets {
			if fire.Before(rec.EnabledAt) {
				continue
			}

			rows = append(rows, &Delivery{
				ReminderID: c.ID, DueAt: c.DueRaw, ChannelID: rec.ID, Status: statusPending,
				NextAttemptAt: now, FireAt: fire, CreatedAt: now,
			})
		}
	}

	return insertDeliveries(ctx, d.DB, rows)
}

func (d *Dispatcher) drainChannel(ctx context.Context, rec *Record) {
	for used := 0; used < perChannelTick && ctx.Err() == nil; {
		now := d.Now().UTC().Truncate(time.Second)
		lease := now.Add(d.SendTimeout + leaseSlack)

		rows, err := claim(ctx, d.DB, rec.ID, now, lease, channelWidth)
		if err != nil {
			slog.Error("notify: claim", "channel_id", rec.ID, "error", err)
			return
		}

		if len(rows) == 0 {
			return
		}

		used += len(rows)

		var wg sync.WaitGroup

		for _, row := range rows {
			wg.Add(1)

			go func() {
				defer wg.Done()

				d.deliver(ctx, rec, row)
			}()
		}

		wg.Wait()
	}
}

func (d *Dispatcher) deliver(ctx context.Context, rec *Record, row claimed) {
	// Sends and result writes outlive a shutdown signal so no outcome is lost.
	bg := context.WithoutCancel(ctx)

	m, err := loadLive(bg, d.DB, row)
	if errors.Is(err, sql.ErrNoRows) {
		if err := cancelOne(bg, d.DB, row); err != nil {
			slog.Error("notify: cancel", "channel_id", rec.ID, "error", err)
		}

		return
	}

	if err != nil {
		slog.Error("notify: load reminder", "channel_id", rec.ID, "error", err)
		return
	}

	loc, _ := d.Schedule(bg)

	sendCtx, cancel := context.WithTimeout(bg, d.SendTimeout)
	defer cancel()

	se := d.send(sendCtx, rec, Message{
		ReminderID: row.ReminderID, Title: m.Title, Notes: m.Notes, PersonName: m.PersonName, DueAt: m.DueDate,
		Location: loc,
	})

	now := d.Now().UTC().Truncate(time.Second)

	status, next := statusSent, now

	if se != nil {
		slog.Warn("notify: delivery failed", "channel_id", rec.ID, "kind", rec.Kind,
			"class", se.Class, "status", se.Status, "attempt", row.Attempts)

		if se.Permanent || row.Attempts >= maxAttempts {
			status = statusFailed
		} else {
			status, next = statusPending, now.Add(Backoff(row.Attempts, se))
		}
	}

	if err := finish(bg, d.DB, row, status, se, next, now); err != nil {
		slog.Error("notify: record result", "channel_id", rec.ID, "error", err)
	}

	if se != nil && se.Class == ClassGone {
		if err := d.Channels.Delete(bg, rec.ID); err != nil && !errors.Is(err, ErrChannelNotFound) {
			slog.Error("notify: prune gone channel", "channel_id", rec.ID, "error", err)
		}
	}
}

func (d *Dispatcher) send(ctx context.Context, rec *Record, m Message) *SendError {
	return SendNow(ctx, d.Registry, d.Client, rec, m)
}

// SendNow is the single send path (dispatcher and test-send). It never returns raw
// adapter text, converts a panic into a permanent failure, and abandons an adapter
// that ignores ctx at the deadline so one stuck channel cannot stall the caller.
func SendNow(ctx context.Context, reg *Registry, client *HTTPClient, rec *Record, m Message) *SendError {
	ch, ok := reg.Get(rec.Kind)
	if !ok || rec.Unreadable {
		return &SendError{Class: ClassInvalidConfig, Permanent: true}
	}

	// The buffered channel lets an abandoned goroutine exit whenever it finally returns.
	res := make(chan *SendError, 1)

	go func() {
		defer func() {
			if recover() != nil {
				slog.Error("notify: channel panic", "channel_id", rec.ID, "kind", rec.Kind)

				res <- &SendError{Class: ClassInternal, Permanent: true}
			}
		}()

		res <- SanitizeErr(ch.Send(ctx, client, rec.Config, m))
	}()

	select {
	case se := <-res:
		return se
	case <-ctx.Done():
		return &SendError{Class: ClassTimeout}
	}
}
