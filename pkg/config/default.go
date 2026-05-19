package config

import "time"

// configDefaults holds default values for all config fields.
// Nested struct fields use dot-separated koanf tag paths (e.g. "SENTRY.DSN").
// Overwrite for env.go EnvConfigMap struct fields.
var configDefaults = map[string]any{
	"DEBUG":      false,
	"SENTRY.DSN": "",
	"TOKEN_AUTH": "",

	// Database
	"DB_PATH":         "data/kith.db",
	"DB_AUTO_MIGRATE": true,

	// File Storage
	"AVATAR_STORAGE_PATH": "data/avatars",
	"GIFT_STORAGE_PATH":   "data/gifts",

	// Auth — SESSION_SECRET must be set in production via environment (≥32 bytes)
	"SESSION_SECRET":    "",
	"APP_PASSWORD_HASH": "",
	"BEHIND_TLS":        false,
	"SESSION_LIFETIME":  30 * 24 * time.Hour,
}
