package people

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/uptrace/bun"

	"github.com/nhymxu/kith-pms/internal/audit"
)

const defaultPageSize = 50

type ListParams struct {
	Query        string
	Page         int
	PageSize     int
	LabelIDs     []int64 // AND-semantics: person must have ALL listed labels
	Sort         string  // sort parameter: name, -name, last_contact, -last_contact
	HasJournal   bool    // when true, only return people linked to at least one journal entry
	FavoriteOnly bool    // when true, only return favorited people
}

type Service struct {
	DB           *bun.DB
	People       PersonRepo
	Contacts     ContactRepo
	Locations    LocationRepo
	FileService  FileService
	LabelsSvc    LabelLoader            // optional; nil = no label loading
	Audit        *audit.Service         // optional; nil = no audit logging
	BirthdaySync BirthdayReminderSyncer // optional; nil = no birthday sync
}

type LabelLoader interface {
	ListByPersonIDs(ctx context.Context, personIDs []int64) (map[int64][]Label, error)
}

// BirthdayReminderSyncer is satisfied by an adapter over reminders.Service (handler layer).
type BirthdayReminderSyncer interface {
	SyncBirthdayRemindersForPerson(ctx context.Context, personID int64, newDOB *DateOnly) error
}

type FileService interface {
	SaveAvatar(personID int64, file multipart.File, header *multipart.FileHeader) (path string, err error)
	DeleteAvatar(personID int64, path string) error
}

func NewService(db *bun.DB) *Service {
	return &Service{
		DB:        db,
		People:    NewPersonRepo(db),
		Contacts:  NewContactRepo(db),
		Locations: NewLocationRepo(db),
	}
}

func (s *Service) Create(ctx context.Context, p Person, contacts []ContactInfo, locations []Location) (int64, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("people: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	id, err := s.People.Create(ctx, tx, p)
	if err != nil {
		return 0, err
	}

	if err := s.Contacts.ReplaceAll(ctx, tx, id, contacts); err != nil {
		return 0, err
	}

	if err := s.Locations.ReplaceAll(ctx, tx, id, locations); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("people: commit create: %w", err)
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityPerson, id, p.Name, audit.ActionCreate)
	}

	return id, nil
}

func (s *Service) Update(ctx context.Context, p Person, contacts []ContactInfo, locations []Location) error {
	var meta *audit.Metadata

	if s.Audit != nil {
		if old, err := s.People.Get(ctx, p.ID); err == nil && old != nil {
			changes := diffPersonFields(*old, p)
			meta = &audit.Metadata{DetailAction: "profile_update", Changes: changes}
		}
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("people: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.People.Update(ctx, tx, p); err != nil {
		return err
	}

	if err := s.Contacts.ReplaceAll(ctx, tx, p.ID, contacts); err != nil {
		return err
	}

	if err := s.Locations.ReplaceAll(ctx, tx, p.ID, locations); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("people: commit update: %w", err)
	}

	if s.Audit != nil {
		if meta != nil {
			s.Audit.Log(ctx, audit.EntityPerson, p.ID, p.Name, audit.ActionUpdate, *meta)
		} else {
			s.Audit.Log(ctx, audit.EntityPerson, p.ID, p.Name, audit.ActionUpdate)
		}
	}

	if s.BirthdaySync != nil {
		_ = s.BirthdaySync.SyncBirthdayRemindersForPerson(ctx, p.ID, p.DateOfBirth)
	}

	return nil
}

func (s *Service) Get(ctx context.Context, id int64) (*Person, error) {
	p, err := s.People.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if p == nil {
		return nil, nil
	}

	contacts, err := s.Contacts.ListByPerson(ctx, id)
	if err != nil {
		return nil, err
	}

	locations, err := s.Locations.ListByPerson(ctx, id)
	if err != nil {
		return nil, err
	}

	if contacts == nil {
		contacts = []ContactInfo{}
	}

	if locations == nil {
		locations = []Location{}
	}

	p.Contacts = contacts
	p.Locations = locations

	if s.LabelsSvc != nil {
		labelsMap, err := s.LabelsSvc.ListByPersonIDs(ctx, []int64{id})
		if err != nil {
			return nil, fmt.Errorf("load labels: %w", err)
		}

		if lbls, ok := labelsMap[id]; ok {
			p.Labels = lbls
		} else {
			p.Labels = []Label{}
		}
	}

	return p, nil
}

func (s *Service) List(ctx context.Context, params ListParams) (*PersonList, error) {
	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}

	if pageSize > 500 {
		pageSize = 500
	}

	page := params.Page
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * pageSize

	total, err := s.People.Count(ctx, params.Query, params.LabelIDs, params.HasJournal, params.FavoriteOnly)
	if err != nil {
		return nil, err
	}

	items, err := s.People.List(
		ctx, params.Query, params.LabelIDs, params.HasJournal, params.FavoriteOnly, pageSize, offset, params.Sort,
	)
	if err != nil {
		return nil, err
	}

	if items == nil {
		items = []Person{}
	}

	// Batch-load labels for all people
	if s.LabelsSvc != nil && len(items) > 0 {
		personIDs := make([]int64, len(items))
		for i, p := range items {
			personIDs[i] = p.ID
		}

		labelsMap, err := s.LabelsSvc.ListByPersonIDs(ctx, personIDs)
		if err != nil {
			return nil, fmt.Errorf("batch load labels: %w", err)
		}

		for i := range items {
			if labels, ok := labelsMap[items[i].ID]; ok {
				items[i].Labels = labels
			}
		}
	}

	return &PersonList{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	var name string

	if s.Audit != nil {
		if p, err := s.People.Get(ctx, id); err == nil && p != nil {
			name = p.Name
		}
	}

	if err := s.People.Delete(ctx, id); err != nil {
		return err
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityPerson, id, name, audit.ActionDelete)
	}

	return nil
}

func (s *Service) GetSelf(ctx context.Context) (*Person, error) {
	return s.People.GetSelf(ctx)
}

func (s *Service) SetSelf(ctx context.Context, personID int64) error {
	person, err := s.People.Get(ctx, personID)
	if err != nil {
		return err
	}

	if person == nil {
		return fmt.Errorf("person not found")
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("people: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.People.ClearSelf(ctx, tx); err != nil {
		return err
	}

	if err := s.People.SetSelf(ctx, tx, personID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("people: commit set self: %w", err)
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityPerson, personID, person.Name, audit.ActionUpdate,
			audit.Metadata{DetailAction: "set_self"})
	}

	return nil
}

func (s *Service) SetFavorite(ctx context.Context, personID int64, favorite bool) error {
	person, err := s.People.Get(ctx, personID)
	if err != nil {
		return err
	}

	if person == nil {
		return fmt.Errorf("person not found")
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("people: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.People.SetFavorite(ctx, tx, personID, favorite); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("people: commit set favorite: %w", err)
	}

	if s.Audit != nil {
		action := "favorite_unset"
		if favorite {
			action = "favorite_set"
		}

		s.Audit.Log(ctx, audit.EntityPerson, personID, person.Name, audit.ActionUpdate,
			audit.Metadata{DetailAction: action})
	}

	return nil
}

// UploadAvatar sets a new avatar for the person.
// If the person already has an avatar, the old file is deleted after the transaction commits.
func (s *Service) UploadAvatar(
	ctx context.Context,
	personID int64,
	file multipart.File,
	header *multipart.FileHeader,
) error {
	if s.FileService == nil {
		return fmt.Errorf("file service not configured")
	}

	person, err := s.People.Get(ctx, personID)
	if err != nil {
		return fmt.Errorf("get person: %w", err)
	}

	if person == nil {
		return fmt.Errorf("person not found")
	}

	oldAvatarPath := person.AvatarPath

	path, err := s.FileService.SaveAvatar(personID, file, header)
	if err != nil {
		return fmt.Errorf("save avatar file: %w", err)
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		_ = s.FileService.DeleteAvatar(personID, path)
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	size := header.Size

	if err := s.People.UpdateAvatar(ctx, tx, personID, path, size); err != nil {
		_ = s.FileService.DeleteAvatar(personID, path)
		return err
	}

	if err := tx.Commit(); err != nil {
		_ = s.FileService.DeleteAvatar(personID, path)
		return fmt.Errorf("commit: %w", err)
	}

	if oldAvatarPath != "" && oldAvatarPath != path {
		_ = s.FileService.DeleteAvatar(personID, oldAvatarPath)
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityPerson, personID, person.Name, audit.ActionUpdate,
			audit.Metadata{DetailAction: "avatar_upload"})
	}

	return nil
}

func (s *Service) DeleteAvatar(ctx context.Context, personID int64) error {
	if s.FileService == nil {
		return fmt.Errorf("file service not configured")
	}

	person, err := s.People.Get(ctx, personID)
	if err != nil {
		return fmt.Errorf("get person: %w", err)
	}

	if person == nil {
		return fmt.Errorf("person not found")
	}

	avatarPath := person.AvatarPath

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.People.ClearAvatar(ctx, tx, personID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	if avatarPath != "" {
		_ = s.FileService.DeleteAvatar(personID, avatarPath)
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityPerson, personID, person.Name, audit.ActionUpdate,
			audit.Metadata{DetailAction: "avatar_delete"})
	}

	return nil
}

// ValidatePeopleExist returns the subset of ids that do NOT exist in the person table.
func (s *Service) ValidatePeopleExist(ctx context.Context, ids []int64) ([]int64, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	var found []int64

	err := s.DB.NewSelect().
		TableExpr("person").
		ColumnExpr("id").
		Where("id IN (?)", bun.List(ids)).
		Scan(ctx, &found)
	if err != nil {
		return nil, fmt.Errorf("people: validate exist: %w", err)
	}

	foundSet := make(map[int64]struct{}, len(found))
	for _, id := range found {
		foundSet[id] = struct{}{}
	}

	var missing []int64

	for _, id := range ids {
		if _, ok := foundSet[id]; !ok {
			missing = append(missing, id)
		}
	}

	return missing, nil
}

func (s *Service) UpdateLastContact(ctx context.Context, personID int64, contactTime time.Time) error {
	person, err := s.People.Get(ctx, personID)
	if err != nil {
		return fmt.Errorf("get person: %w", err)
	}

	if person == nil {
		return fmt.Errorf("person not found")
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.People.UpdateLastContact(ctx, tx, personID, contactTime); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	if s.Audit != nil {
		var oldTS, newTS string
		if person.LastContactAt != nil {
			oldTS = person.LastContactAt.UTC().Format(time.RFC3339)
		}

		newTS = contactTime.UTC().Format(time.RFC3339)
		s.Audit.Log(ctx, audit.EntityPerson, personID, person.Name, audit.ActionUpdate,
			audit.Metadata{
				DetailAction: "last_contact_update",
				Changes:      []audit.Change{{Field: "last_contact_at", Old: oldTS, New: newTS}},
			})
	}

	return nil
}
