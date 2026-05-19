package settings

// UserSettings holds the user's display preferences.
type UserSettings struct {
	DateFormat string `json:"date_format"`
	TimeFormat string `json:"time_format"`
	Timezone   string `json:"timezone"`
}

var Defaults = UserSettings{
	DateFormat: "YYYY-MM-DD",
	TimeFormat: "24h",
	Timezone:   "UTC",
}

const (
	KeyDateFormat = "date_format"
	KeyTimeFormat = "time_format"
	KeyTimezone   = "timezone"
)
