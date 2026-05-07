package gifts

import (
	"context"
	"database/sql"
	"fmt"
	"mime/multipart"

	"github.com/nhymxu/kith-pms/internal/audit"
	"github.com/nhymxu/kith-pms/internal/files"
)

type Service struct {
	db      *sql.DB
	repo    *Repo
	Audit   *audit.Service
	FileSvc files.FileService
}

func NewService(db *sql.DB) *Service {
	return &Service{
		db:   db,
		repo: NewRepo(db),
	}
}

func (s *Service) Create(ctx context.Context, g *Gift) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	id, err := s.repo.Create(ctx, tx, g)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityGift, id, g.Title, audit.ActionCreate)
	}

	return id, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*Gift, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, params ListParams) ([]GiftWithPerson, error) {
	return s.repo.List(ctx, params)
}

func (s *Service) Update(ctx context.Context, g *Gift) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.repo.Update(ctx, tx, g); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityGift, g.ID, g.Title, audit.ActionUpdate)
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	var (
		title     string
		imagePath string
	)

	if g, err := s.repo.GetByID(ctx, id); err == nil && g != nil {
		title = g.Title
		imagePath = g.ImagePath
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

	if s.FileSvc != nil && imagePath != "" {
		_ = s.FileSvc.DeleteGiftImage(id, imagePath)
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityGift, id, title, audit.ActionDelete)
	}

	return nil
}

func (s *Service) UploadImage(
	ctx context.Context,
	giftID int64,
	file multipart.File,
	header *multipart.FileHeader,
) error {
	if s.FileSvc == nil {
		return fmt.Errorf("file service not configured")
	}

	// Capture old path before overwriting so we can delete the orphaned file.
	var oldPath string
	if existing, err := s.repo.GetByID(ctx, giftID); err == nil && existing != nil {
		oldPath = existing.ImagePath
	}

	path, err := s.FileSvc.SaveGiftImage(giftID, file, header)
	if err != nil {
		return fmt.Errorf("save gift image: %w", err)
	}

	mimeType := header.Header.Get("Content-Type")

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.repo.UpdateImage(ctx, tx, giftID, path, mimeType); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	if s.FileSvc != nil && oldPath != "" && oldPath != path {
		_ = s.FileSvc.DeleteGiftImage(giftID, oldPath)
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityGift, giftID, "", audit.ActionUpdate)
	}

	return nil
}

func (s *Service) DeleteImage(ctx context.Context, giftID int64) error {
	g, err := s.repo.GetByID(ctx, giftID)
	if err != nil {
		return err
	}

	oldPath := g.ImagePath

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.repo.UpdateImage(ctx, tx, giftID, "", ""); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	if s.FileSvc != nil && oldPath != "" {
		_ = s.FileSvc.DeleteGiftImage(giftID, oldPath)
	}

	if s.Audit != nil {
		s.Audit.Log(ctx, audit.EntityGift, giftID, g.Title, audit.ActionUpdate)
	}

	return nil
}
