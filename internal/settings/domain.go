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
	NetworkOnlyMineDepth   string `json:"network_only_mine_depth"` // "direct" or "alter"

	AllowFavoriteToggleOnList bool   `json:"allow_favorite_toggle_on_list"`
	FavoriteFirstDefault      bool   `json:"favorite_first_default"`
	DefaultPeopleSort         string `json:"default_people_sort"`
	DefaultPageSize           int    `json:"default_page_size"`

	DashboardFavoritesCount   int `json:"dashboard_favorites_count"`
	DashboardLastContactCount int `json:"dashboard_last_contact_count"`

	Theme string `json:"theme"`

	NavLayout string `json:"nav_layout"`

	// NumberFormat controls the grouping/decimal separators for displayed
	// numbers, e.g. "1,234.56" or "1.234,56".
	NumberFormat string `json:"number_format"`

	SearchScope []string `json:"search_scope"`
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
	NetworkOnlyMineDepth:   "direct",

	AllowFavoriteToggleOnList: true,
	FavoriteFirstDefault:      false,
	DefaultPeopleSort:         "name",
	DefaultPageSize:           25,

	DashboardFavoritesCount:   5,
	DashboardLastContactCount: 5,

	Theme: "quiet-ink",

	NavLayout: "top",

	NumberFormat: "1,234.56",

	SearchScope: []string{"people", "journal", "gifts", "notes"},
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
	KeyNetworkOnlyMineDepth   = "network_only_mine_depth"

	KeyAllowFavoriteToggleOnList = "allow_favorite_toggle_on_list"
	KeyFavoriteFirstDefault      = "favorite_first_default"
	KeyDefaultPeopleSort         = "default_people_sort"
	KeyDefaultPageSize           = "default_page_size"

	KeyDashboardFavoritesCount   = "dashboard_favorites_count"
	KeyDashboardLastContactCount = "dashboard_last_contact_count"

	KeyTheme = "theme"

	KeyNavLayout = "nav_layout"

	KeyNumberFormat = "number_format"

	KeySearchScope = "search_scope"
)
