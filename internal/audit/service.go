package audit

import (
	"context"
	"log/slog"
	"time"

	"github.com/uptrace/bun"
)

// Service provides audit logging for domain mutations.
type Service struct {
	db   *bun.DB
	repo *Repo
}

func NewService(db *bun.DB) *Service {
	return &Service{db: db, repo: NewRepo()}
}

// Log writes an audit entry best-effort: errors are logged as warnings but
// never returned so a logging failure never breaks the primary operation.
// An optional Metadata value may be passed as the last argument.
func (s *Service) Log(
	ctx context.Context,
	entityType EntityType,
	entityID int64,
	entityName string,
	action Action,
	meta ...Metadata,
) {
	var m *Metadata
	if len(meta) > 0 {
		m = &meta[0]
	}

	e := Entry{
		EntityType: entityType,
		EntityID:   entityID,
		EntityName: entityName,
		Action:     action,
		ActorID:    ActorFromCtx(ctx),
		Metadata:   m,
		CreatedAt:  time.Now().UTC(),
	}
	if err := s.repo.Insert(ctx, s.db, e); err != nil {
		slog.WarnContext(ctx, "audit log failed", "err", err,
			"entity_type", entityType, "entity_id", entityID, "action", action)
	}
}

func (s *Service) List(ctx context.Context, p ListParams) ([]Entry, error) {
	return s.repo.List(ctx, s.db, p)
}

// Purge deletes entries older than days. days=0 is a no-op.
func (s *Service) Purge(ctx context.Context, days int) (int64, error) {
	if days <= 0 {
		return 0, nil
	}

	return s.repo.Purge(ctx, s.db, days)
}
