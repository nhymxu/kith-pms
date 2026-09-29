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

export const noteWithPersonSchema = noteSchema.extend({
	person_name: z.string(),
	person_nickname: z.string().optional().default(""),
	person_has_avatar: z.boolean(),
});

export const noteAllListSchema = noteListSchema.extend({
	items: z.array(noteWithPersonSchema),
});

export type NoteWithPerson = z.infer<typeof noteWithPersonSchema>;
export type NoteAllList = z.infer<typeof noteAllListSchema>;
export type Note = z.infer<typeof noteSchema>;
export type NoteList = z.infer<typeof noteListSchema>;

export const noteRequestSchema = z.object({
	title: z.string().optional().default(""),
	content: z.string().min(1),
});

export type NoteRequest = z.infer<typeof noteRequestSchema>;
