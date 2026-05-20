package settings

// UserSettings holds the user's display preferences.
type UserSettings struct {
	DateFormat            string `json:"date_format"`
	TimeFormat            string `json:"time_format"`
	Timezone              string `json:"timezone"`
	AuditLogRetentionDays int    `json:"audit_log_retention_days"` // 0 = disabled
}

var Defaults = UserSettings{
	DateFormat:            "YYYY-MM-DD",
	TimeFormat:            "24h",
	Timezone:              "UTC",
	AuditLogRetentionDays: 0,
}

const (
	KeyDateFormat            = "date_format"
	KeyTimeFormat            = "time_format"
	KeyTimezone              = "timezone"
	KeyAuditLogRetentionDays = "audit_log_retention_days"
)
