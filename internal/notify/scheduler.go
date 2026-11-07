package notify

import (
	"context"
	"time"

	"github.com/nhymxu/kith-pms/internal/settings"
)

// SettingsSchedule reads the user's timezone and notify hour from settings.
func SettingsSchedule(svc *settings.Service) Schedule {
	return func(ctx context.Context) (*time.Location, int) {
		cfg, err := svc.Get(ctx)
		if err != nil {
			return time.UTC, settings.Defaults.NotifyHour
		}

		loc, err := time.LoadLocation(cfg.Timezone)
		if err != nil {
			return time.UTC, settings.Defaults.NotifyHour
		}

		return loc, cfg.NotifyHour
	}
}
