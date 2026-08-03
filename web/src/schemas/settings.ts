import { z } from "zod";

export const userSettingsSchema = z.object({
	date_format: z.enum(["YYYY-MM-DD", "MM/DD/YYYY", "DD/MM/YYYY"]),
	time_format: z.enum(["24h", "12h"]),
	timezone: z.string().min(1),
	theme: z
		.enum([
			"quiet-ink",
			"warm-album",
			"bold-press",
			"nightdesk",
			"softclay",
			"ledger",
		])
		.default("quiet-ink"),
	audit_log_retention_days: z.number().int().min(0).default(0),
	network_color_by: z.enum(["labels", "type"]).default("labels"),
	network_show_avatar: z.boolean().default(false),
	network_show_only_mine: z.boolean().default(false),
	network_show_unconnected: z.boolean().default(true),
	network_only_mine_depth: z.enum(["direct", "alter"]).default("direct"),
	allow_favorite_toggle_on_list: z.boolean().default(true),
	favorite_first_default: z.boolean().default(false),
	default_people_sort: z
		.enum(["name", "-name", "last_contact", "-last_contact"])
		.default("name"),
	default_page_size: z.number().int().min(10).max(200).default(25),
	dashboard_favorites_count: z.number().int().min(1).max(20).default(5),
	dashboard_last_contact_count: z.number().int().min(1).max(20).default(5),
	nav_layout: z.enum(["top", "side"]).default("top"),
	search_scope: z
		.array(z.enum(["people", "journal", "gifts", "notes"]))
		.min(1)
		.default(["people", "journal", "gifts", "notes"]),
});

export type UserSettings = z.infer<typeof userSettingsSchema>;

// GET/PUT /v1/settings return the persisted UserSettings plus this read-only,
// server-config-derived field — it's not a user preference, so it's kept out
// of userSettingsSchema (the PUT request body shape).
export const settingsResponseSchema = userSettingsSchema.extend({
	max_upload_size_mb: z.number().int().positive(),
});

export type SettingsResponse = z.infer<typeof settingsResponseSchema>;
