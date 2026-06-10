import { z } from "zod";

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
		.min(1, "Name is required")
		.max(64, "Name must be 64 chars or fewer"),
	color: z
		.string()
		.regex(/^#[0-9a-fA-F]{6}$/, "Color must be a hex color e.g. #a1b2c3"),
});

export type JournalLabel = z.infer<typeof journalLabelSchema>;
export type JournalLabelList = z.infer<typeof journalLabelListSchema>;
export type JournalLabelRequest = z.infer<typeof journalLabelRequestSchema>;
