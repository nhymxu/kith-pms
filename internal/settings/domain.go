package settings

// UserSettings holds the user's display preferences.
type UserSettings struct {
	DateFormat             string `json:"date_format"`
	TimeFormat             string `json:"time_format"`
	Timezone               string `json:"timezone"`
	AuditLogRetentionDays  int    `json:"audit_log_retention_days"` // 0 = disabled
	NetworkColorBy         string `json:"network_color_by"`
	NetworkShowAvatar      bool   `json:"network_show_avatar"`
	NetworkShowOnlyMine    bool   `json:"network_show_only_mine"`
	NetworkShowUnconnected bool   `json:"network_show_unconnected"`
}

var Defaults = UserSettings{
	DateFormat:             "YYYY-MM-DD",
	TimeFormat:             "24h",
	Timezone:               "UTC",
	AuditLogRetentionDays:  0,
	NetworkColorBy:         "labels",
	NetworkShowAvatar:      false,
	NetworkShowOnlyMine:    false,
	NetworkShowUnconnected: true,
}

const (
	KeyDateFormat             = "date_format"
	KeyTimeFormat             = "time_format"
	KeyTimezone               = "timezone"
	KeyAuditLogRetentionDays  = "audit_log_retention_days"
	KeyNetworkColorBy         = "network_color_by"
	KeyNetworkShowAvatar      = "network_show_avatar"
	KeyNetworkShowOnlyMine    = "network_show_only_mine"
	KeyNetworkShowUnconnected = "network_show_unconnected"
)
