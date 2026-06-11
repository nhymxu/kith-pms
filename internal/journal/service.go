package journal

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/uptrace/bun"

	"github.com/nhymxu/kith-pms/internal/audit"
)

const defaultPageSize = 30

// ListParams holds filter and pagination parameters for listing activities.
type ListParams struct {
	Query           string
	PersonIDs       []int64
	LabelIDs        []int64 // filters by people-label (activities whose linked people have these labels)
	JournalLabelIDs []int64 // filters by journal-label (OR within; AND with other filters)
	FromDate        string  // "YYYY-MM-DD"
	ToDate          string  // "YYYY-MM-DD"
	Page            int
	PageSize        int // default 30
}

// Service provides business logic for managing journal activities.
type Service struct {
	DB         *bun.DB
	Activities ActivityRepo
	Links      ActivityPersonRepo
	Labels     LabelRepo
	LabelLinks LabelAssignmentRepo
	Audit      *audit.Service // optional; nil = no audit logging
	PeopleSvc  PeopleServiceInterface
}

// PeopleServiceInterface defines methods needed from people.Service.
type PeopleServiceInterface interface {
	GetSelf(ctx context.Context) (*PersonAdapter, error)
	Get(ctx context.Context, id int64) (*PersonAdapter, error)
	UpdateLastContact(ctx context.Context, personID int64, contactTime time.Time) error
}

// PersonAdapter wraps person data for interface compatibility.
type PersonAdapter struct {
	PersonID      int64
	LastContactAt *time.Time
}

// NewService constructs a Service wired to db.
func NewService(db *bun.DB) *Service {
	return &Service{
		DB:         db,
		Activities: NewActivityRepo(db),
		Links:      NewActivityPersonRepo(db),
		Labels:     NewLabelRepo(db),
		LabelLinks: NewLabelAssignmentRepo(db),
	}
}

// Create inserts a new activity and links people + labels in a single transaction.
func (s *Service) Create(ctx context.Context, a Activity, personIDs []int64, labelIDs []int64) (int64, error) {
	// Validate label IDs before opening the transaction to avoid SQLite read-inside-write deadlock.
	validLabelIDs, err := s.filterLabelIDs(ctx, labelIDs)
	if err != nil {
		return 0, err
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("journal: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	id, err := s.Activities.Create(ctx, tx, a)
	if err != nil {
		return 0, err
	}

	if err := s.Links.ReplaceAll(ctx, tx, id, personIDs); err != nil {
		return 0, err
	}

	if s.LabelLinks != nil {
		if err := s.LabelLinks.ReplaceAll(ctx, tx, id, validLabelIDs); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("journal: commit create: %w", err)
	}

	// Update last contact after transaction commits to avoid nested transaction deadlock.
	a.ID = id
	if err := s.updateLastContactForParticipants(ctx, a, personIDs); err != nil {
		// Log error but don't fail the entire operation since the journal entry was created successfully.
		slog.Warn("failed to update last contact for participants", "activity_id", id, "error", err)
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityJournal, id, a.Title, audit.ActionCreate)
	}

	return id, nil
}

// Update replaces an activity's fields and all person links + labels in a single transaction.
func (s *Service) Update(ctx context.Context, a Activity, personIDs []int64, labelIDs []int64) error {
	// Validate label IDs before opening the transaction to avoid SQLite read-inside-write deadlock.
	validLabelIDs, err := s.filterLabelIDs(ctx, labelIDs)
	if err != nil {
		return err
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("journal: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.Activities.Update(ctx, tx, a); err != nil {
		return err
	}

	if err := s.Links.ReplaceAll(ctx, tx, a.ID, personIDs); err != nil {
		return err
	}

	if s.LabelLinks != nil {
		if err := s.LabelLinks.ReplaceAll(ctx, tx, a.ID, validLabelIDs); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("journal: commit update: %w", err)
	}

	// Update last contact after transaction commits to avoid nested transaction deadlock.
	if err := s.updateLastContactForParticipants(ctx, a, personIDs); err != nil {
		// Log error but don't fail the entire operation since the journal entry was updated successfully.
		slog.Warn("failed to update last contact for participants", "activity_id", a.ID, "error", err)
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityJournal, a.ID, a.Title, audit.ActionUpdate)
	}

	return nil
}

// updateLastContactForParticipants updates last_contact_at for people linked to this activity,
// but only if the activity includes the self person and the activity date is newer.
func (s *Service) updateLastContactForParticipants(ctx context.Context, a Activity, personIDs []int64) error {
	if s.PeopleSvc == nil || len(personIDs) == 0 {
		return nil
	}

	selfPerson, err := s.PeopleSvc.GetSelf(ctx)
	if err != nil || selfPerson == nil {
		return nil
	}

	selfID := selfPerson.PersonID
	hasSelf := false

	for _, pid := range personIDs {
		if pid == selfID {
			hasSelf = true
			break
		}
	}

	if !hasSelf {
		return nil
	}

	var oatStr string
	if a.OccurredAtTime != nil {
		oatStr = *a.OccurredAtTime
	}

	activityTime, err := parseActivityTimestamp(a.OccurredAtDate, oatStr)
	if err != nil {
		return nil
	}

	for _, pid := range personIDs {
		if pid == selfID {
			continue
		}

		person, err := s.PeopleSvc.Get(ctx, pid)
		if err != nil || person == nil {
			continue
		}

		if person.LastContactAt == nil || activityTime.After(*person.LastContactAt) {
			if err := s.PeopleSvc.UpdateLastContact(ctx, pid, activityTime); err != nil {
				slog.Warn("failed to update last contact", "person_id", pid, "error", err)
			}
		}
	}

	return nil
}

// filterLabelIDs returns only label IDs that exist in journal_label (drops unknown ones).
// Returns nil when Labels repo is not wired or input is empty.
func (s *Service) filterLabelIDs(ctx context.Context, ids []int64) ([]int64, error) {
	if s.Labels == nil || len(ids) == 0 {
		return nil, nil
	}

	return s.Labels.FilterExisting(ctx, ids)
}

// parseActivityTimestamp converts occurred_at_date + occurred_at_time to time.Time.
// If time is empty, uses midnight UTC.
func parseActivityTimestamp(date, timeStr string) (time.Time, error) {
	if timeStr != "" {
		combined := date + " " + timeStr

		t, err := time.Parse("2006-01-02 15:04", combined)
		if err != nil {
			return time.Time{}, err
		}

		return t.UTC(), nil
	}

	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
}

// Get returns the activity with the given id, or nil, nil when not found.
func (s *Service) Get(ctx context.Context, id int64) (*Activity, error) {
	a, err := s.Activities.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if a == nil {
		return nil, nil
	}

	people, err := s.Links.ListByActivity(ctx, id)
	if err != nil {
		return nil, err
	}

	if people == nil {
		people = []ActivityPerson{}
	}

	a.People = people

	if s.LabelLinks != nil {
		labels, err := s.LabelLinks.ListByActivityID(ctx, id)
		if err != nil {
			return nil, err
		}

		if labels == nil {
			labels = []Label{}
		}

		a.Labels = labels
	} else {
		a.Labels = []Label{}
	}

	return a, nil
}

// Delete removes an activity; FTS mirror is updated by the activity_ad trigger.
func (s *Service) Delete(ctx context.Context, id int64) error {
	var title string

	if s.Audit != nil {
		if a, err := s.Activities.Get(ctx, id); err == nil && a != nil {
			title = a.Title
		}
	}

	if err := s.Activities.Delete(ctx, id); err != nil {
		return err
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityJournal, id, title, audit.ActionDelete)
	}

	return nil
}

func (s *Service) List(ctx context.Context, params ListParams) (*ActivityList, error) {
	if params.PageSize <= 0 {
		params.PageSize = defaultPageSize
	}

	if params.PageSize > 500 {
		params.PageSize = 500
	}

	if params.Page < 1 {
		params.Page = 1
	}

	total, err := s.Activities.Count(ctx, params)
	if err != nil {
		return nil, err
	}

	items, err := s.Activities.List(ctx, params)
	if err != nil {
		return nil, err
	}

	if items == nil {
		items = []Activity{}
	}

	// Batch-load people for all activities.
	for i := range items {
		people, err := s.Links.ListByActivity(ctx, items[i].ID)
		if err != nil {
			return nil, err
		}

		if people == nil {
			people = []ActivityPerson{}
		}

		items[i].People = people
	}

	// Batch-load journal labels (single query, no N+1).
	if s.LabelLinks != nil && len(items) > 0 {
		ids := make([]int64, len(items))
		for i, item := range items {
			ids[i] = item.ID
		}

		labelMap, err := s.LabelLinks.ListByActivityIDs(ctx, ids)
		if err != nil {
			return nil, err
		}

		for i := range items {
			labels := labelMap[items[i].ID]
			if labels == nil {
				labels = []Label{}
			}

			items[i].Labels = labels
		}
	} else {
		for i := range items {
			items[i].Labels = []Label{}
		}
	}

	return &ActivityList{
		Items:    items,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
	}, nil
}
