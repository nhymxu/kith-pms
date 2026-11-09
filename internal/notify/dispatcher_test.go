package notify_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/uptrace/bun"

	"github.com/nhymxu/kith-pms/internal/notify"
	"github.com/nhymxu/kith-pms/internal/notify/notifytest"
	"github.com/nhymxu/kith-pms/internal/testutil"
)

// recChannel behaves per cfg.mode and records every Send by cfg.name.
type recChannel struct {
	mu    sync.Mutex
	calls map[string][]int64
	gate  chan struct{}

	lastLoc *time.Location
}

func (c *recChannel) Kind() string           { return "rec" }
func (c *recChannel) SecretFields() []string { return []string{"token"} }
func (c *recChannel) PublicFields() []string { return []string{"name", "mode"} }
func (c *recChannel) Validate(json.RawMessage) error {
	return nil
}

func (c *recChannel) count(name string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.calls[name])
}

func (c *recChannel) Send(ctx context.Context, _ *notify.HTTPClient, cfg json.RawMessage, m notify.Message) error {
	var v struct{ Name, Mode string }

	_ = json.Unmarshal(cfg, &v)

	c.mu.Lock()
	c.calls[v.Name] = append(c.calls[v.Name], m.ReminderID)
	c.lastLoc = m.Location
	n := len(c.calls[v.Name])
	c.mu.Unlock()

	switch v.Mode {
	case "hang":
		<-ctx.Done()
		return ctx.Err()
	case "deaf":
		time.Sleep(time.Second) // ignores ctx on purpose
	case "panic":
		panic(notifytest.Canary)
	case "gone":
		return &notify.SendError{Class: notify.ClassGone, Status: 410, Permanent: true}
	case "permanent":
		return &notify.SendError{Class: notify.ClassHTTP4xx, Status: 400, Permanent: true}
	case "flaky":
		if n <= 2 {
			return &notify.SendError{Class: notify.ClassHTTP5xx, Status: 503}
		}
	case "leak":
		return errors.New(notifytest.Canary)
	case "gate":
		<-c.gate
	}

	return nil
}

type env struct {
	t     *testing.T
	db    *bun.DB
	repo  *notify.ChannelRepo
	ch    *recChannel
	d     *notify.Dispatcher
	clock time.Time
	loc   *time.Location
	hour  int
}

func newEnv(t *testing.T) *env {
	t.Helper()

	e := &env{
		t: t, db: testutil.NewDB(t), loc: time.UTC, hour: 9,
		clock: time.Now().UTC().Truncate(time.Second),
		ch:    &recChannel{calls: map[string][]int64{}, gate: make(chan struct{})},
	}
	e.repo = notify.NewChannelRepo(e.db, newBox(t))

	reg := notify.NewRegistry()
	reg.Register(e.ch)

	e.d = notify.NewDispatcher(e.db, reg, e.repo, notify.NewHTTPClient(),
		func(context.Context) (*time.Location, int) { return e.loc, e.hour })
	e.d.Now = func() time.Time { return e.clock }
	e.d.SendTimeout = 100 * time.Millisecond

	return e
}

func (e *env) channel(name, mode string) int64 {
	e.t.Helper()

	b, _ := json.Marshal(map[string]string{"name": name, "mode": mode, "token": "tok"})

	rec, err := e.repo.Create(context.Background(), "rec", name, true, b)
	if err != nil {
		e.t.Fatal(err)
	}

	e.exec(`UPDATE notification_channel SET enabled_at = ? WHERE id = ?`, e.clock.Add(-72*time.Hour), rec.ID)

	return rec.ID
}

func (e *env) exec(q string, args ...any) {
	e.t.Helper()

	if _, err := e.db.NewRaw(q, args...).Exec(context.Background()); err != nil {
		e.t.Fatal(err)
	}
}

func (e *env) reminder(due time.Time) int64 {
	e.t.Helper()

	return e.reminderFor(due, nil)
}

func (e *env) reminderFor(due time.Time, person *int64) int64 {
	e.t.Helper()

	var id int64

	err := e.db.NewRaw(`INSERT INTO reminder (title, notes, due_date, person_id, created_at, updated_at)
VALUES ('t', 'n', ?, ?, ?, ?) RETURNING id`, due, person, due.Add(-24*time.Hour), due.Add(-24*time.Hour)).
		Scan(context.Background(), &id)
	if err != nil {
		e.t.Fatal(err)
	}

	return id
}

func (e *env) person() int64 {
	e.t.Helper()

	var id int64
	if err := e.db.NewRaw(`INSERT INTO person (name) VALUES ('P') RETURNING id`).
		Scan(context.Background(), &id); err != nil {
		e.t.Fatal(err)
	}

	return id
}

type row struct {
	Status   string `bun:"status"`
	Attempts int    `bun:"attempts"`
	Class    string `bun:"last_error_class"`
}

func (e *env) rows(reminderID int64) []row {
	e.t.Helper()

	var out []row

	err := e.db.NewRaw(`SELECT status, attempts, last_error_class FROM reminder_delivery
WHERE reminder_id = ?`, reminderID).Scan(context.Background(), &out)
	if err != nil {
		e.t.Fatal(err)
	}

	return out
}

func (e *env) tick() { e.d.Tick(context.Background()) }

func (e *env) want(reminderID int64, status string, attempts int) {
	e.t.Helper()

	r := e.rows(reminderID)
	if len(r) != 1 || r[0].Status != status || r[0].Attempts != attempts {
		e.t.Fatalf("reminder %d: want %s/%d, got %+v", reminderID, status, attempts, r)
	}
}

func TestFanOutSendsOncePerChannel(t *testing.T) {
	e := newEnv(t)
	e.channel("a", "ok")
	e.channel("b", "ok")
	e.channel("c", "ok")

	id := e.reminder(e.clock.Add(-time.Hour))

	e.tick()
	e.tick()

	for _, n := range []string{"a", "b", "c"} {
		if e.ch.count(n) != 1 {
			t.Errorf("channel %s sent %d times", n, e.ch.count(n))
		}
	}

	if len(e.rows(id)) != 3 {
		t.Fatalf("want 3 delivery rows, got %d", len(e.rows(id)))
	}
}

func TestRetryThenSuccessWithExactBackoff(t *testing.T) {
	e := newEnv(t)
	e.channel("f", "flaky")
	id := e.reminder(e.clock.Add(-time.Hour))

	e.tick()
	e.want(id, "pending", 1)

	var next time.Time
	if err := e.db.NewRaw(`SELECT next_attempt_at FROM reminder_delivery WHERE reminder_id = ?`, id).
		Scan(context.Background(), &next); err != nil {
		t.Fatal(err)
	}

	if !next.Equal(e.clock.Add(time.Minute)) {
		t.Fatalf("first backoff: want %v got %v", e.clock.Add(time.Minute), next)
	}

	e.tick() // not due yet
	e.want(id, "pending", 1)

	e.clock = e.clock.Add(time.Minute)
	e.tick()
	e.want(id, "pending", 2)

	e.clock = e.clock.Add(5 * time.Minute)
	e.tick()
	e.want(id, "sent", 3)
}

func TestBackoffSchedule(t *testing.T) {
	want := []time.Duration{time.Minute, 5 * time.Minute, 30 * time.Minute, 2 * time.Hour, 2 * time.Hour}
	for i, w := range want {
		if got := notify.Backoff(i+1, &notify.SendError{}); got != w {
			t.Errorf("attempt %d: want %v got %v", i+1, w, got)
		}
	}

	if got := notify.Backoff(1, &notify.SendError{RetryAfter: 7 * time.Minute}); got != 7*time.Minute {
		t.Errorf("Retry-After must win, got %v", got)
	}
}

func TestPermanentAndPanicFailAtOnce(t *testing.T) {
	e := newEnv(t)
	e.channel("p", "permanent")
	e.channel("x", "panic")

	id := e.reminder(e.clock.Add(-time.Hour))
	e.tick()

	for _, r := range e.rows(id) {
		if r.Status != "failed" || r.Attempts != 1 {
			t.Fatalf("want failed/1, got %+v", r)
		}
	}
}

func TestNotSentCases(t *testing.T) {
	e := newEnv(t)
	e.channel("a", "ok")
	due := e.clock.Add(-time.Hour)

	completed := e.reminder(due)
	e.exec(`UPDATE reminder SET completed = 1 WHERE id = ?`, completed)

	deletedP, archivedP := e.person(), e.person()
	e.exec(`UPDATE person SET deleted_at = ? WHERE id = ?`, e.clock, deletedP)
	e.exec(`UPDATE person SET archived_at = ? WHERE id = ?`, e.clock, archivedP)

	soft := e.reminderFor(due, &deletedP)
	arch := e.reminderFor(due, &archivedP)
	stale := e.reminder(e.clock.Add(-25 * time.Hour))

	alreadyDue := e.reminder(due)
	e.exec(`UPDATE reminder SET created_at = ? WHERE id = ?`, due.Add(time.Minute), alreadyDue)

	e.tick()

	cases := map[string]int64{
		"completed": completed, "softdeleted": soft, "archived": arch, "stale": stale, "created>=due": alreadyDue,
	}

	for name, id := range cases {
		if r := e.rows(id); len(r) != 0 {
			t.Errorf("%s: unexpected delivery rows %+v", name, r)
		}
	}

	if e.ch.count("a") != 0 {
		t.Fatalf("nothing should send, got %d", e.ch.count("a"))
	}
}

func TestCompleteOrEditWhilePendingCancels(t *testing.T) {
	e := newEnv(t)
	e.channel("f", "flaky")

	done, edited := e.reminder(e.clock.Add(-time.Hour)), e.reminder(e.clock.Add(-2*time.Hour))
	e.tick()

	e.exec(`UPDATE reminder SET completed = 1 WHERE id = ?`, done)
	e.exec(`UPDATE reminder SET due_date = ? WHERE id = ?`, e.clock.Add(3*time.Hour), edited)

	before := e.ch.count("f")
	e.clock = e.clock.Add(time.Minute)
	e.tick()

	if e.ch.count("f") != before {
		t.Fatal("cancelled reminders must not retry")
	}

	e.want(done, "cancelled", 1)
	e.want(edited, "cancelled", 1)
}

func TestHangingChannelDoesNotBlockOthers(t *testing.T) {
	e := newEnv(t)
	e.channel("slow", "hang")
	e.channel("fast", "ok")

	ids := make([]int64, 20)
	for i := range ids {
		ids[i] = e.reminder(e.clock.Add(-time.Hour))
	}

	e.tick()
	e.tick()

	if e.ch.count("fast") != 20 {
		t.Fatalf("healthy channel sent %d/20", e.ch.count("fast"))
	}

	// the hanging channel is backing off after its first try, never re-sent within the same minute
	if got := e.ch.count("slow"); got != 20 {
		t.Fatalf("hanging channel must try each row exactly once, got %d", got)
	}
}

func TestDisableEnableAndNewChannelDoNotReplay(t *testing.T) {
	e := newEnv(t)
	id := e.channel("a", "flaky")
	rem := e.reminder(e.clock.Add(-time.Hour))

	e.tick()
	e.want(rem, "pending", 1)

	ctx := context.Background()
	if err := e.repo.SetEnabled(ctx, id, false); err != nil {
		t.Fatal(err)
	}

	e.want(rem, "cancelled", 1)

	e.clock = e.clock.Add(time.Hour)
	e.tick()

	if e.ch.count("a") != 1 {
		t.Fatal("disabled channel must not send")
	}

	// re-enable: enabled_at is now, so the old fire time is not replayed
	time.Sleep(10 * time.Millisecond)

	if err := e.repo.SetEnabled(ctx, id, true); err != nil {
		t.Fatal(err)
	}

	e.clock = time.Now().UTC().Add(time.Second)
	e.tick()

	if e.ch.count("a") != 1 {
		t.Fatal("re-enabled channel replayed an old reminder")
	}

	if err := e.repo.Delete(ctx, id); err != nil {
		t.Fatal(err)
	}

	if len(e.rows(rem)) != 0 {
		t.Fatal("channel delete must remove its deliveries")
	}
}

func TestRestartResumesPendingRow(t *testing.T) {
	e := newEnv(t)
	e.channel("f", "flaky")
	id := e.reminder(e.clock.Add(-time.Hour))

	e.tick()
	e.want(id, "pending", 1)

	e.clock = e.clock.Add(time.Minute)

	reg := notify.NewRegistry()
	reg.Register(e.ch)

	d2 := notify.NewDispatcher(e.db, reg, e.repo, notify.NewHTTPClient(),
		func(context.Context) (*time.Location, int) { return e.loc, e.hour })
	d2.Now = func() time.Time { return e.clock }
	d2.Tick(context.Background())

	e.want(id, "pending", 2)
}

func TestCancelledContextMidSendStillRecordsSent(t *testing.T) {
	e := newEnv(t)
	e.d.SendTimeout = 5 * time.Second
	e.channel("g", "gate")
	id := e.reminder(e.clock.Add(-time.Hour))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		e.d.Tick(ctx)
		close(done)
	}()

	for i := 0; i < 200 && e.ch.count("g") == 0; i++ {
		time.Sleep(10 * time.Millisecond)
	}

	cancel()
	close(e.ch.gate)
	<-done

	e.want(id, "sent", 1)
}

func TestDateOnlyFiresAtNotifyHourLocal(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		t.Skip("tzdata missing")
	}

	due := time.Date(2030, 5, 10, 0, 0, 0, 0, time.UTC)
	if got := notify.FireTime(due, loc, 9); !got.Equal(time.Date(2030, 5, 10, 2, 0, 0, 0, time.UTC)) {
		t.Fatalf("09:00 ICT is 02:00 UTC, got %v", got)
	}

	timed := due.Add(7*time.Hour + time.Minute)
	if got := notify.FireTime(timed, loc, 9); !got.Equal(timed) {
		t.Fatalf("timed reminder must keep its instant, got %v", got)
	}

	e := newEnv(t)
	e.loc = loc
	e.channel("a", "ok")
	e.clock = time.Date(2030, 5, 10, 1, 0, 0, 0, time.UTC) // 08:00 ICT
	e.exec(`INSERT INTO reminder (title, due_date, created_at, updated_at) VALUES ('b', ?, ?, ?)`,
		due, due.Add(-48*time.Hour), due.Add(-48*time.Hour))
	e.exec(`UPDATE notification_channel SET enabled_at = ?`, e.clock.Add(-72*time.Hour))

	e.tick()

	if e.ch.count("a") != 0 {
		t.Fatal("must wait for notify hour")
	}

	e.clock = e.clock.Add(61 * time.Minute) // 09:01 ICT
	e.tick()

	if e.ch.count("a") != 1 {
		t.Fatalf("must fire after notify hour, sent %d", e.ch.count("a"))
	}
}

func TestSecretsAbsentFromStoredErrorsAndLogs(t *testing.T) {
	var buf bytes.Buffer

	old := slog.Default()

	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(old) })

	e := newEnv(t)
	e.channel("l", "leak")
	e.channel("x", "panic")
	e.reminder(e.clock.Add(-time.Hour))
	e.tick()

	var dump []map[string]any
	if err := e.db.NewRaw(`SELECT * FROM reminder_delivery`).Scan(context.Background(), &dump); err != nil {
		t.Fatal(err)
	}

	out, _ := json.Marshal(dump)
	notifytest.AssertNoCanary(t, "delivery table", string(out))
	notifytest.AssertNoCanary(t, "logs", buf.String())

	if !strings.Contains(string(out), notify.ClassInternal) {
		t.Fatalf("expected class internal in %s", out)
	}
}

func TestAdapterIgnoringContextDoesNotStallTick(t *testing.T) {
	e := newEnv(t)
	e.channel("d", "deaf")
	e.channel("ok", "ok")
	e.reminder(e.clock.Add(-time.Hour))

	start := time.Now()

	e.tick()

	if e.ch.count("ok") != 1 {
		t.Fatal("healthy channel must send")
	}

	if time.Since(start) > 900*time.Millisecond {
		t.Fatalf("tick waited for a deaf adapter: %v", time.Since(start))
	}
}

func TestRowsForDisabledChannelAreNotSentAfterReenable(t *testing.T) {
	e := newEnv(t)
	id := e.channel("a", "ok")
	rem := e.reminder(e.clock.Add(-time.Hour))

	// simulate the enqueue-vs-disable race: a pending row exists for a channel that
	// was disabled and re-enabled afterwards (enabled_at moved past fire_at)
	e.exec(`INSERT INTO reminder_delivery (reminder_id, due_at, channel_id, next_attempt_at, fire_at, created_at)
SELECT id, due_date, ?, ?, ?, ? FROM reminder WHERE id = ?`, id, e.clock, e.clock.Add(-time.Hour), e.clock, rem)
	e.exec(`UPDATE notification_channel SET enabled_at = ?`, e.clock)

	e.d.Schedule = func(context.Context) (*time.Location, int) { return e.loc, e.hour }
	e.tick()

	if e.ch.count("a") != 0 {
		t.Fatal("row older than enabled_at must not send")
	}
}

func TestFireTimeTable(t *testing.T) {
	ny, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("tzdata missing")
	}

	cases := []struct {
		name string
		due  time.Time
		loc  *time.Location
		hour int
		want time.Time
	}{
		{
			"date-only utc",
			time.Date(2030, 1, 5, 0, 0, 0, 0, time.UTC),
			time.UTC,
			9,
			time.Date(2030, 1, 5, 9, 0, 0, 0, time.UTC),
		},
		{
			"date-only ny winter",
			time.Date(2030, 1, 5, 0, 0, 0, 0, time.UTC),
			ny,
			9,
			time.Date(2030, 1, 5, 14, 0, 0, 0, time.UTC),
		},
		{
			"date-only ny summer",
			time.Date(2030, 7, 5, 0, 0, 0, 0, time.UTC),
			ny,
			9,
			time.Date(2030, 7, 5, 13, 0, 0, 0, time.UTC),
		},
		{"timed", time.Date(2030, 1, 5, 0, 0, 1, 0, time.UTC), ny, 9, time.Date(2030, 1, 5, 0, 0, 1, 0, time.UTC)},
	}

	for _, c := range cases {
		if got := notify.FireTime(c.due, c.loc, c.hour); !got.Equal(c.want) {
			t.Errorf("%s: want %v got %v", c.name, c.want, got)
		}
	}
}

func TestFractionalSecondBoundaryOrdering(t *testing.T) {
	e := newEnv(t)
	e.channel("a", "ok")

	// next_attempt_at stored as whole seconds must be claimable exactly at that second,
	// and a due time with fractional seconds must not be claimed a moment early
	early := e.clock.Add(-time.Hour)
	id := e.reminder(early)
	e.tick()
	e.want(id, "sent", 1)

	flaky := newEnv(t)
	flaky.channel("f", "flaky")
	rid := flaky.reminder(flaky.clock.Add(-time.Hour))
	flaky.tick()

	flaky.clock = flaky.clock.Add(time.Minute - time.Second).Add(999 * time.Millisecond)
	flaky.tick()
	flaky.want(rid, "pending", 1)

	flaky.clock = flaky.clock.Add(time.Second)
	flaky.tick()
	flaky.want(rid, "pending", 2)
}

func TestDateOnlyCreatedEarlierOnLocalDueDayStillFires(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		t.Skip("tzdata missing")
	}

	e := newEnv(t)
	e.loc = loc
	e.channel("a", "ok")

	// 08:00 ICT on May 10 = 01:00 UTC, after the stored 00:00 UTC due date
	created := time.Date(2030, 5, 10, 1, 0, 0, 0, time.UTC)
	due := time.Date(2030, 5, 10, 0, 0, 0, 0, time.UTC)
	e.clock = created.Add(2 * time.Hour) // 10:00 ICT, past the 09:00 fire time
	e.exec(
		`INSERT INTO reminder (title, due_date, created_at, updated_at) VALUES ('v', ?, ?, ?)`,
		due,
		created,
		created,
	)
	e.exec(`UPDATE notification_channel SET enabled_at = ?`, e.clock.Add(-72*time.Hour))
	e.tick()

	if e.ch.count("a") != 1 {
		t.Fatalf("reminder created before its local fire time must send, got %d", e.ch.count("a"))
	}

	if e.ch.lastLoc == nil || e.ch.lastLoc.String() != loc.String() {
		t.Fatalf("message must carry the user's location, got %v", e.ch.lastLoc)
	}
}

func TestDateOnlyCreatedAfterFireTimeNeverSends(t *testing.T) {
	e := newEnv(t)
	e.channel("a", "ok")

	due := time.Date(2030, 5, 10, 0, 0, 0, 0, time.UTC)
	e.clock = due.Add(11 * time.Hour)
	created := due.Add(10 * time.Hour) // after the 09:00 fire time: import stamping "today"
	e.exec(
		`INSERT INTO reminder (title, due_date, created_at, updated_at) VALUES ('i', ?, ?, ?)`,
		due,
		created,
		created,
	)
	e.exec(`UPDATE notification_channel SET enabled_at = ?`, e.clock.Add(-72*time.Hour))
	e.tick()

	if e.ch.count("a") != 0 {
		t.Fatal("reminder created after its fire time must not send")
	}
}

func TestGoneChannelIsPruned(t *testing.T) {
	e := newEnv(t)
	gone := e.channel("g", "gone")
	e.channel("ok", "ok")

	id := e.reminder(e.clock.Add(-time.Hour))
	e.tick()

	if _, err := e.repo.Get(context.Background(), gone); !errors.Is(err, notify.ErrChannelNotFound) {
		t.Fatalf("gone channel should be deleted, got %v", err)
	}

	if rows := e.rows(id); len(rows) != 1 || rows[0].Status != "sent" {
		t.Fatalf("other channel unaffected, got %+v", rows)
	}
}

func TestVAPIDStoreStableAndRotateDeletesWebPush(t *testing.T) {
	db := testutil.NewDB(t)
	box := newBox(t)
	store := notify.NewVAPIDStore(db, box)
	repo := notify.NewChannelRepo(db, box)
	ctx := context.Background()

	priv1, pub1, err := store.Keys(ctx)
	if err != nil || priv1 == "" || pub1 == "" {
		t.Fatalf("keys: %v", err)
	}

	if priv2, pub2, _ := store.Keys(ctx); priv2 != priv1 || pub2 != pub1 {
		t.Fatal("keys must be stable across calls")
	}

	var stored string
	if err := db.NewRaw(`SELECT CAST(private_enc AS TEXT) FROM webpush_vapid`).Scan(ctx, &stored); err != nil ||
		strings.Contains(stored, priv1) {
		t.Fatalf("private key must be sealed at rest (err %v)", err)
	}

	wp, _ := repo.Create(ctx, notify.KindWebPush, "b", true, json.RawMessage(`{}`))
	other, _ := repo.Create(ctx, "fake", "o", true, json.RawMessage(`{}`))

	pub3, err := store.Rotate(ctx)
	if err != nil || pub3 == pub1 {
		t.Fatalf("rotate: %v", err)
	}

	if _, err := repo.Get(ctx, wp.ID); !errors.Is(err, notify.ErrChannelNotFound) {
		t.Fatal("webpush channel must be deleted on rotate")
	}

	if _, err := repo.Get(ctx, other.ID); err != nil {
		t.Fatal("other kinds must survive rotate")
	}

	if _, pub4, _ := store.Keys(ctx); pub4 != pub3 {
		t.Fatal("rotated key must persist")
	}
}
