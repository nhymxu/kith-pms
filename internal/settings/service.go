package settings

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/uptrace/bun"
)

var (
	ErrInvalidDateFormat    = errors.New("settings: date_format must be one of YYYY-MM-DD, MM/DD/YYYY, DD/MM/YYYY")
	ErrInvalidTimeFormat    = errors.New("settings: time_format must be one of 24h, 12h")
	ErrInvalidTimezone      = errors.New("settings: timezone must not be empty")
	ErrInvalidRetentionDays = errors.New("settings: audit_log_retention_days must be >= 0")
)

var validDateFormats = map[string]bool{
	"YYYY-MM-DD": true,
	"MM/DD/YYYY": true,
	"DD/MM/YYYY": true,
}

var validTimeFormats = map[string]bool{
	"24h": true,
	"12h": true,
}

type Service struct {
	Repo Repo
}

func NewService(db *bun.DB) *Service {
	return &Service{Repo: NewRepo(db)}
}

func (s *Service) Get(ctx context.Context) (UserSettings, error) {
	rows, err := s.Repo.GetAll(ctx)
	if err != nil {
		return Defaults, err
	}

	result := Defaults
	if v, ok := rows[KeyDateFormat]; ok {
		result.DateFormat = v
	}

	if v, ok := rows[KeyTimeFormat]; ok {
		result.TimeFormat = v
	}

	if v, ok := rows[KeyTimezone]; ok {
		result.Timezone = v
	}

	if v, ok := rows[KeyAuditLogRetentionDays]; ok {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			result.AuditLogRetentionDays = n
		}
	}

	return result, nil
}

func (s *Service) Update(ctx context.Context, in UserSettings) (UserSettings, error) {
	if !validDateFormats[in.DateFormat] {
		return UserSettings{}, ErrInvalidDateFormat
	}

	if !validTimeFormats[in.TimeFormat] {
		return UserSettings{}, ErrInvalidTimeFormat
	}

	if in.Timezone == "" {
		return UserSettings{}, ErrInvalidTimezone
	}

	if in.AuditLogRetentionDays < 0 {
		return UserSettings{}, ErrInvalidRetentionDays
	}

	now := time.Now().UTC()
	for key, val := range map[string]string{
		KeyDateFormat:            in.DateFormat,
		KeyTimeFormat:            in.TimeFormat,
		KeyTimezone:              in.Timezone,
		KeyAuditLogRetentionDays: strconv.Itoa(in.AuditLogRetentionDays),
	} {
		if err := s.Repo.Set(ctx, key, val, now); err != nil {
			return UserSettings{}, err
		}
	}

	return in, nil
}

func (s *Service) GetRetentionDays(ctx context.Context) (int, error) {
	cfg, err := s.Get(ctx)
	if err != nil {
		return 0, err
	}

	return cfg.AuditLogRetentionDays, nil
}
