package api

import (
	"context"
	"time"

	"github.com/nhymxu/kith-pms/internal/journal"
	"github.com/nhymxu/kith-pms/internal/people"
)

type JournalPeopleAdapter struct {
	Svc *people.Service
}

func (a *JournalPeopleAdapter) GetSelf(ctx context.Context) (*journal.PersonAdapter, error) {
	p, err := a.Svc.GetSelf(ctx)
	if err != nil {
		return nil, err
	}

	if p == nil {
		return nil, nil
	}

	return &journal.PersonAdapter{
		PersonID:      p.ID,
		LastContactAt: p.LastContactAt,
	}, nil
}

func (a *JournalPeopleAdapter) Get(ctx context.Context, id int64) (*journal.PersonAdapter, error) {
	p, err := a.Svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if p == nil {
		return nil, nil
	}

	return &journal.PersonAdapter{
		PersonID:      p.ID,
		LastContactAt: p.LastContactAt,
	}, nil
}

func (a *JournalPeopleAdapter) UpdateLastContact(ctx context.Context, personID int64, contactTime time.Time) error {
	return a.Svc.UpdateLastContact(ctx, personID, contactTime)
}
