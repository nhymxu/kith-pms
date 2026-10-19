import { z } from "zod";
import i18n from "#/lib/i18n";

export const labelSchema = z.object({
	id: z.number(),
	name: z.string(),
	color: z.string(),
	created_at: z.string(),
	count: z.number().optional().default(0),
});

export const labelListSchema = z.array(labelSchema);

export type Label = z.infer<typeof labelSchema>;
export type LabelList = z.infer<typeof labelListSchema>;

export const labelRequestSchema = z.object({
	name: z
		.string()
		.min(1, { error: () => i18n.t("errors:validation.nameRequired") })
		.max(64, {
			error: () => i18n.t("errors:validation.nameTooLong", { max: 64 }),
		}),
	color: z
		.string()
		.regex(/^#[0-9a-fA-F]{6}$/, "Must be a valid hex color e.g. #a1b2c3"),
});

export type LabelRequest = z.infer<typeof labelRequestSchema>;
