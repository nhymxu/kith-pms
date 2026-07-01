package config

import (
	"time"

	"github.com/jinzhu/copier"
	"github.com/nhymxu/gommon/cfgloader"
)

// Config define mapping struct field and environment field
type Config struct {
	Debug  bool `koanf:"DEBUG"`
	Sentry struct {
		DSN string `koanf:"DSN"`
	} `koanf:"SENTRY"`

	TokenAuth string `koanf:"TOKEN_AUTH" copier:"-"`

	// Database
	DBPath         string `koanf:"DB_PATH"`
	DBAutoMigrate  bool   `koanf:"DB_AUTO_MIGRATE"`
	DBMaxOpenConns int    `koanf:"DB_MAX_OPEN_CONNS"`

	// File Storage
	AvatarStoragePath string `koanf:"AVATAR_STORAGE_PATH"`
	GiftStoragePath   string `koanf:"GIFT_STORAGE_PATH"`

	// Auth
	SessionSecret   string        `koanf:"SESSION_SECRET" copier:"-"`
	AppPasswordHash string        `koanf:"APP_PASSWORD_HASH" copier:"-"`
	BehindTLS       bool          `koanf:"BEHIND_TLS"`
	SessionLifetime time.Duration `koanf:"SESSION_LIFETIME"`
}

func (c *Config) Sanitized() Config {
	var cc Config

	// Secrets excluded ❌
	err := copier.Copy(&cc, &c)
	if err != nil {
		return Config{}
	}

	return cc
}

// configDefaults holds default values for all config fields.
// Nested struct fields use dot-separated koanf tag paths (e.g. "SENTRY.DSN").
var configDefaults = map[string]any{
	"DEBUG":      false,
	"SENTRY.DSN": "",
	"TOKEN_AUTH": "",

	// Database
	"DB_PATH":           "data/kith.db",
	"DB_AUTO_MIGRATE":   true,
	"DB_MAX_OPEN_CONNS": 1,

	// File Storage
	"AVATAR_STORAGE_PATH": "data/avatars",
	"GIFT_STORAGE_PATH":   "data/gifts",

	// Auth — SESSION_SECRET must be set in production via environment (≥32 bytes)
	"SESSION_SECRET":    "",
	"APP_PASSWORD_HASH": "",
	"BEHIND_TLS":        false,
	"SESSION_LIFETIME":  30 * 24 * time.Hour,
}

// C is the global config instance.
var C Config

// Load reads env file and populates C.
func Load(cfgFile string) error {
	var err error

	C, err = cfgloader.LoadConfig[Config](cfgFile, configDefaults)

	return err
}
