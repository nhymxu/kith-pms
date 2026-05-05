package audit

import (
	"context"
	"database/sql"
	"log/slog"
)

// Service provides audit logging for domain mutations.
type Service struct {
	db   *sql.DB
	repo *Repo
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db, repo: NewRepo()}
}

// Log writes an audit entry best-effort: errors are logged as warnings but
// never returned so a logging failure never breaks the primary operation.
func (s *Service) Log(ctx context.Context, entityType EntityType, entityID int64, entityName string, action Action) {
	e := Entry{
		EntityType: entityType,
		EntityID:   entityID,
		EntityName: entityName,
		Action:     action,
		ActorID:    ActorFromCtx(ctx),
	}
	if err := s.repo.Insert(ctx, s.db, e); err != nil {
		slog.WarnContext(ctx, "audit log failed", "err", err,
			"entity_type", entityType, "entity_id", entityID, "action", action)
	}
}

// List returns paginated audit entries.
func (s *Service) List(ctx context.Context, p ListParams) ([]Entry, error) {
	return s.repo.List(ctx, s.db, p)
}
