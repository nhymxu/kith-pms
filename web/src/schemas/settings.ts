import { z } from "zod";

export const userSettingsSchema = z.object({
	date_format: z.enum(["YYYY-MM-DD", "MM/DD/YYYY", "DD/MM/YYYY"]),
	time_format: z.enum(["24h", "12h"]),
	timezone: z.string().min(1),
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
});

export type UserSettings = z.infer<typeof userSettingsSchema>;
