package reminders

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"

	"github.com/nhymxu/kith-pms/internal/audit"
	"github.com/nhymxu/kith-pms/internal/people"
)

// JournalLastContacter is satisfied by journal.Service — used for relative_contact recurrence.
type JournalLastContacter interface {
	LastContactDate(ctx context.Context, personID int64) (time.Time, error)
}

// PersonDOBLookup is satisfied by an adapter over people.Service (handler layer).
type PersonDOBLookup interface {
	PersonDOB(ctx context.Context, personID int64) (dob *people.DateOnly, name string, err error)
}

var (
	ErrNoPerson = errors.New("person not found")
	ErrNoDOB    = errors.New("person has no date of birth")
)

type Service struct {
	db        *bun.DB
	repo      *Repo
	Audit     *audit.Service       // optional; nil = no audit logging
	Journal   JournalLastContacter // optional; nil = fallback to completedAt
	PersonDOB PersonDOBLookup      // optional; nil = birthday features no-op
}

func NewService(db *bun.DB) *Service {
	return &Service{
		db:   db,
		repo: NewRepo(db),
	}
}

func (s *Service) Create(ctx context.Context, rem *Reminder) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	id, err := s.repo.Create(ctx, tx, rem)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityReminder, id, rem.Title, audit.ActionCreate)
	}

	return id, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*Reminder, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, rem *Reminder) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.repo.Update(ctx, tx, rem); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityReminder, rem.ID, rem.Title, audit.ActionUpdate)
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	var title string

	if s.Audit != nil {
		if r, err := s.repo.GetByID(ctx, id); err == nil && r != nil {
			title = r.Title
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.repo.Delete(ctx, tx, id); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityReminder, id, title, audit.ActionDelete)
	}

	return nil
}

func (s *Service) List(ctx context.Context, params ListParams) ([]ReminderWithPerson, error) {
	return s.repo.List(ctx, params)
}

func (s *Service) GetUpcoming(ctx context.Context, days int) ([]ReminderWithPerson, error) {
	return s.repo.ListUpcoming(ctx, days)
}

func (s *Service) GetOverdue(ctx context.Context) ([]ReminderWithPerson, error) {
	return s.repo.ListOverdue(ctx)
}

func (s *Service) MarkComplete(ctx context.Context, id int64) error {
	rem, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get reminder: %w", err)
	}

	now := time.Now()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.repo.MarkComplete(ctx, tx, id, now); err != nil {
		return err
	}

	if rem.IsBirthday() {
		if s.PersonDOB != nil && rem.PersonID != nil {
			dob, _, err := s.PersonDOB.PersonDOB(ctx, *rem.PersonID)
			if err == nil && dob != nil {
				daysBefore := 0
				if rem.RecurrenceRule.DaysBeforeDob != nil {
					daysBefore = *rem.RecurrenceRule.DaysBeforeDob
				}

				nextDue := NextBirthdayDueAfter(*dob, daysBefore, rem.DueDate)
				next := &Reminder{
					Title:          rem.Title,
					Notes:          rem.Notes,
					PersonID:       rem.PersonID,
					RecurrenceRule: rem.RecurrenceRule,
					DueDate:        nextDue,
				}

				if _, err := s.repo.Create(ctx, tx, next); err != nil {
					return fmt.Errorf("spawn next birthday: %w", err)
				}
			}
			// dob == nil (cleared) → reminder just completes, no spawn
		}
	} else if rem.RecurrenceRule != nil {
		var lastContact time.Time

		if s.Journal != nil && rem.PersonID != nil {
			lastContact, _ = s.Journal.LastContactDate(ctx, *rem.PersonID)
		}

		rem.CompletedAt = &now
		nextDue := computeNextDue(rem, lastContact)

		if !nextDue.IsZero() {
			next := &Reminder{
				Title:             rem.Title,
				Notes:             rem.Notes,
				PersonID:          rem.PersonID,
				ImportantDateID:   rem.ImportantDateID,
				RecurrenceRule:    rem.RecurrenceRule,
				RecurrenceEndDate: rem.RecurrenceEndDate,
				DueDate:           nextDue,
			}

			if _, err := s.repo.Create(ctx, tx, next); err != nil {
				return fmt.Errorf("spawn next occurrence: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityReminder, id, rem.Title, audit.ActionUpdate)
	}

	return nil
}

func (s *Service) CountByStatus(ctx context.Context, status string) (int, error) {
	return s.repo.CountByStatus(ctx, status)
}

// SyncBirthdayRemindersForPerson recomputes due_date for all PENDING birthday reminders
// of a person, or deletes ALL of them if newDOB is nil.
func (s *Service) SyncBirthdayRemindersForPerson(ctx context.Context, personID int64, newDOB *people.DateOnly) error {
	// Query reminders before beginning transaction to avoid lock issues.
	rems, err := s.repo.FindBirthdayRemindersByPersonID(ctx, personID)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if newDOB == nil {
		if err := s.repo.DeleteBirthdayRemindersByPersonID(ctx, tx, personID); err != nil {
			return err
		}

		return tx.Commit()
	}

	now := time.Now()

	for _, rem := range rems {
		daysBefore := 0
		if rem.RecurrenceRule != nil && rem.RecurrenceRule.DaysBeforeDob != nil {
			daysBefore = *rem.RecurrenceRule.DaysBeforeDob
		}

		rem.DueDate = ComputeBirthdayDueDate(*newDOB, daysBefore, now)
		if err := s.repo.Update(ctx, tx, rem); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// EnsureBirthdayReminder creates an on-day (days_before=0) birthday reminder for the person
// if none with days_before=0 exists. Idempotent. Returns nil if feature is off or no DOB set.
func (s *Service) EnsureBirthdayReminder(ctx context.Context, personID int64) error {
	if s.PersonDOB == nil {
		return nil
	}

	dob, name, err := s.PersonDOB.PersonDOB(ctx, personID)
	if err != nil {
		return fmt.Errorf("lookup person dob: %w", err)
	}

	if dob == nil {
		return nil
	}

	rems, err := s.repo.FindBirthdayRemindersByPersonID(ctx, personID)
	if err != nil {
		return err
	}

	for _, rem := range rems {
		if rem.RecurrenceRule == nil || rem.RecurrenceRule.DaysBeforeDob == nil ||
			*rem.RecurrenceRule.DaysBeforeDob == 0 {
			return nil // already exists
		}
	}

	zero := 0
	rem := &Reminder{
		Title:    name + "'s birthday",
		PersonID: &personID,
		RecurrenceRule: &RecurrenceRule{
			Type:          RecurrenceBirthday,
			DaysBeforeDob: &zero,
		},
		DueDate: ComputeBirthdayDueDate(*dob, 0, time.Now()),
	}

	_, err = s.Create(ctx, rem)

	return err
}

// BuildBirthdayReminder resolves DOB and computes due_date for a birthday reminder request.
// Returns a populated *Reminder ready for Create/Update, or ErrNoPerson/ErrNoDOB.
func (s *Service) BuildBirthdayReminder(
	ctx context.Context,
	personID int64,
	daysBefore int,
	title string,
) (*Reminder, error) {
	if s.PersonDOB == nil {
		return nil, fmt.Errorf("birthday feature unavailable")
	}

	if daysBefore < 0 || daysBefore > 365 {
		return nil, fmt.Errorf("days_before_dob must be between 0 and 365")
	}

	dob, name, err := s.PersonDOB.PersonDOB(ctx, personID)
	if err != nil {
		return nil, ErrNoPerson
	}

	if dob == nil {
		return nil, ErrNoDOB
	}

	if title == "" {
		title = name + "'s birthday"
	}

	due := ComputeBirthdayDueDate(*dob, daysBefore, time.Now())

	return &Reminder{
		Title:    title,
		PersonID: &personID,
		RecurrenceRule: &RecurrenceRule{
			Type:          RecurrenceBirthday,
			DaysBeforeDob: &daysBefore,
		},
		DueDate: due,
	}, nil
}

// HasBirthdayReminderForPerson returns true if the person has at least one non-completed birthday reminder.
func (s *Service) HasBirthdayReminderForPerson(ctx context.Context, personID int64) (bool, error) {
	return s.repo.HasBirthdayReminderForPerson(ctx, personID)
}

// DeleteBirthdayRemindersForPerson deletes all birthday reminders for a person.
func (s *Service) DeleteBirthdayRemindersForPerson(ctx context.Context, personID int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.repo.DeleteBirthdayRemindersByPersonID(ctx, tx, personID); err != nil {
		return err
	}

	return tx.Commit()
}
