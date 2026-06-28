package dates

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/uptrace/bun"

	"github.com/nhymxu/kith-pms/internal/audit"
)

type Service struct {
	db    *bun.DB
	repo  ImportantDateRepo
	Audit *audit.Service // optional; nil = no audit logging
}

func NewService(db *bun.DB) *Service {
	return &Service{
		db:   db,
		repo: NewRepo(db),
	}
}

func (s *Service) ListByPerson(ctx context.Context, personID int64) ([]ImportantDate, error) {
	return s.repo.ListByPerson(ctx, personID)
}

func (s *Service) ReplaceForPerson(ctx context.Context, personID int64, dates []ImportantDate) error {
	for _, d := range dates {
		if d.Kind == "birthday" {
			return fmt.Errorf("birthday kind is not allowed in important_date; use person.date_of_birth instead")
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.repo.ReplaceAll(ctx, tx, personID, dates); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityDate, personID, "", audit.ActionUpdate)
	}

	return nil
}

func (s *Service) OnThisDay(ctx context.Context, today time.Time) ([]OnThisDayItem, error) {
	monthDay := today.Format("01-02")
	todayISO := today.Format("2006-01-02")

	return s.repo.OnThisDay(ctx, monthDay, todayISO)
}

func (s *Service) Upcoming(ctx context.Context, today time.Time, days int) ([]OnThisDayItem, error) {
	all, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	type upcomingItem struct {
		item OnThisDayItem
		next time.Time
	}

	var upcoming []upcomingItem

	for _, item := range all {
		next := nextOccurrence(item.Date, today)
		if next.IsZero() {
			continue
		}

		daysDiff := int(next.Sub(today).Hours() / 24)
		if daysDiff <= days {
			// Calculate YearsSince if year-having and recurring
			if !item.Date.IsYearless() && item.Date.Recurring {
				dateVal, err := time.Parse("2006-01-02", item.Date.DateValue)
				if err == nil {
					item.YearsSince = next.Year() - dateVal.Year()
				}
			}

			upcoming = append(upcoming, upcomingItem{item: item, next: next})
		}
	}

	// Sort by next occurrence ascending
	sort.Slice(upcoming, func(i, j int) bool {
		return upcoming[i].next.Before(upcoming[j].next)
	})

	result := make([]OnThisDayItem, len(upcoming))
	for i, u := range upcoming {
		result[i] = u.item
	}

	return result, nil
}
