package api

import (
	"context"
	"log/slog"
	"time"

	"github.com/nhymxu/kith-pms/internal/people"
)

// RunPeoplePurgeGC hard-deletes soft-deleted people past the retention window.
// retentionDays<=0 disables the purge (people.Service.PurgeExpired is a no-op).
func RunPeoplePurgeGC(ctx context.Context, svc *people.Service, retentionDays int) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n, err := svc.PurgeExpired(ctx, retentionDays)
			if err != nil {
				slog.Warn("people purge GC error", "error", err)
			} else {
				slog.Debug("people purge GC: expired people deleted", "count", n)
			}
		}
	}
}
