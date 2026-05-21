package api

import (
	"context"
	"log/slog"
	"time"

	"github.com/nhymxu/kith-pms/internal/auth"
)

func RunSessionGC(ctx context.Context, repo auth.SessionRepo) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := repo.DeleteExpiredSessions(ctx); err != nil {
				slog.Warn("session GC error", "error", err)
			} else {
				slog.Debug("session GC: expired sessions deleted")
			}
		}
	}
}
