import { z } from "zod";

export const noteSchema = z.object({
	id: z.number(),
	person_id: z.number(),
	title: z.string().optional().default(""),
	content: z.string().optional().default(""),
	created_at: z.string(),
	updated_at: z.string(),
});

export const noteListSchema = z.object({
	items: z.array(noteSchema),
	total: z.number(),
	page: z.number(),
	page_size: z.number(),
});

export type Note = z.infer<typeof noteSchema>;
export type NoteList = z.infer<typeof noteListSchema>;

export const noteRequestSchema = z.object({
	title: z.string().optional().default(""),
	content: z.string().min(1),
});

export type NoteRequest = z.infer<typeof noteRequestSchema>;
