package notify

import (
	"context"
	"time"

	"github.com/nhymxu/kith-pms/internal/settings"
)

// SettingsSchedule reads the user's timezone from settings. The notify hour is
// fixed at its default until it becomes a user setting.
func SettingsSchedule(svc *settings.Service) Schedule {
	return func(ctx context.Context) (*time.Location, int) {
		cfg, err := svc.Get(ctx)
		if err != nil {
			return time.UTC, defaultNotify
		}

		loc, err := time.LoadLocation(cfg.Timezone)
		if err != nil {
			return time.UTC, defaultNotify
		}

		return loc, defaultNotify
	}
}
