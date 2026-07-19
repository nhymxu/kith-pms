package note

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/uptrace/bun"

	"github.com/nhymxu/kith-pms/internal/audit"
)

const snippetMaxLen = 60

type Service struct {
	db    *bun.DB
	repo  *Repo
	Audit *audit.Service
}

func NewService(db *bun.DB) *Service {
	return &Service{
		db:   db,
		repo: NewRepo(db),
	}
}

// auditDetail returns the note title, falling back to a snippet of the content
// when the title is empty (notes, unlike activities, allow an empty title).
func auditDetail(n *Note) string {
	if n.Title != "" {
		return n.Title
	}

	return snippet(n.Content)
}

func snippet(content string) string {
	content = strings.TrimSpace(content)
	if utf8.RuneCountInString(content) <= snippetMaxLen {
		return content
	}

	runes := []rune(content)

	return string(runes[:snippetMaxLen]) + "…"
}

func (s *Service) Create(ctx context.Context, n *Note) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	id, err := s.repo.Create(ctx, tx, n)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}

	if s.Audit != nil {
		personName, _ := s.repo.PersonName(ctx, n.PersonID)
		s.Audit.Log(ctx, audit.EntityNote, id, personName, audit.ActionCreate,
			audit.Metadata{Label: auditDetail(n)})
	}

	return id, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*Note, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListByPerson(ctx context.Context, personID int64, page, pageSize int) (*List, error) {
	if page < 1 {
		page = 1
	}

	return s.repo.ListByPerson(ctx, personID, page, pageSize)
}

func (s *Service) Search(ctx context.Context, query string, limit int) ([]WithPerson, error) {
	return s.repo.Search(ctx, query, limit)
}

func (s *Service) Update(ctx context.Context, n *Note) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.repo.Update(ctx, tx, n); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	if s.Audit != nil {
		personName, _ := s.repo.PersonName(ctx, n.PersonID)
		s.Audit.Log(ctx, audit.EntityNote, n.ID, personName, audit.ActionUpdate,
			audit.Metadata{Label: auditDetail(n)})
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
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
		personName, _ := s.repo.PersonName(ctx, n.PersonID)
		s.Audit.Log(ctx, audit.EntityNote, id, personName, audit.ActionDelete,
			audit.Metadata{Label: auditDetail(n)})
	}

	return nil
}
