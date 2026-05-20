import { z } from "zod";

export const userSettingsSchema = z.object({
	date_format: z.enum(["YYYY-MM-DD", "MM/DD/YYYY", "DD/MM/YYYY"]),
	time_format: z.enum(["24h", "12h"]),
	timezone: z.string().min(1),
	audit_log_retention_days: z.number().int().min(0).default(0),
});

export type UserSettings = z.infer<typeof userSettingsSchema>;
