import { z } from "zod";
import i18n from "#/lib/i18n";

export const journalLabelSchema = z.object({
	id: z.number(),
	name: z.string(),
	color: z.string(),
	created_at: z.string(),
	count: z.number().optional().default(0),
});

export const journalLabelListSchema = z.array(journalLabelSchema);

export const journalLabelRequestSchema = z.object({
	name: z
		.string()
		.min(1, { error: () => i18n.t("errors:validation.nameRequired") })
		.max(64, {
			error: () => i18n.t("errors:validation.nameTooLong", { max: 64 }),
		}),
	color: z.string().regex(/^#[0-9a-fA-F]{6}$/, {
		error: () => i18n.t("errors:validation.colorInvalid"),
	}),
});

export type JournalLabel = z.infer<typeof journalLabelSchema>;
export type JournalLabelList = z.infer<typeof journalLabelListSchema>;
export type JournalLabelRequest = z.infer<typeof journalLabelRequestSchema>;
