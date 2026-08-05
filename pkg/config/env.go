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
	MaxUploadSizeMB   int    `koanf:"MAX_UPLOAD_SIZE_MB"`
	ImageMaxEdgePX    int    `koanf:"IMAGE_MAX_EDGE_PX"`
	ImageJPEGQuality  int    `koanf:"IMAGE_JPEG_QUALITY"`

	// Auth
	SessionSecret   string        `koanf:"SESSION_SECRET" copier:"-"`
	AppPasswordHash string        `koanf:"APP_PASSWORD_HASH" copier:"-"`
	BehindTLS       bool          `koanf:"BEHIND_TLS"`
	SessionLifetime time.Duration `koanf:"SESSION_LIFETIME"`
}

// Fallbacks mirroring configDefaults, for callers that construct config
// directly without going through Load (e.g. tests building handlers).
const (
	defaultMaxUploadSizeMB  = 32
	defaultImageMaxEdgePX   = 1600
	defaultImageJPEGQuality = 85
)

// EffectiveMaxUploadBytes returns the configured avatar/gift image upload cap
// in bytes, falling back to the default when unset (e.g. in tests that
// construct handlers directly without calling config.Load).
func (c *Config) EffectiveMaxUploadBytes() int64 {
	mb := c.MaxUploadSizeMB
	if mb <= 0 {
		mb = defaultMaxUploadSizeMB
	}

	return int64(mb) * 1024 * 1024
}

// EffectiveImageMaxEdgePX returns the longest-edge cap applied to cropped image
// uploads by the SPA.
func (c *Config) EffectiveImageMaxEdgePX() int {
	if c.ImageMaxEdgePX <= 0 {
		return defaultImageMaxEdgePX
	}

	return c.ImageMaxEdgePX
}

// EffectiveImageJPEGQuality returns the JPEG quality (1-100) the SPA encodes
// cropped uploads with. Out-of-range values are clamped rather than passed
// through: canvas.toBlob silently ignores an invalid quality argument, which
// would produce confusingly large uploads with no visible error.
func (c *Config) EffectiveImageJPEGQuality() int {
	q := c.ImageJPEGQuality
	if q <= 0 {
		return defaultImageJPEGQuality
	}

	if q > 100 {
		return 100
	}

	return q
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
	"MAX_UPLOAD_SIZE_MB":  defaultMaxUploadSizeMB,
	"IMAGE_MAX_EDGE_PX":   defaultImageMaxEdgePX,
	"IMAGE_JPEG_QUALITY":  defaultImageJPEGQuality,

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
