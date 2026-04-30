package config

// configDefaults holds default values for all config fields.
// Nested struct fields use dot-separated koanf tag paths (e.g. "SENTRY.DSN").
// Overwrite for env.go EnvConfigMap struct fields.
var configDefaults = map[string]any{
	"DEBUG":      false,
	"SENTRY.DSN": "",
	"TOKEN_AUTH": "",
}
