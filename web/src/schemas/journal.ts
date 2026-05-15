import { z } from "zod"

export const activityPersonSchema = z.object({
	person_id: z.number(),
	name: z.string(),
})

export const journalActivitySchema = z.object({
	id: z.number(),
	title: z.string(),
	occurred_at_date: z.string(),
	occurred_at_time: z.string().optional().default(""),
	content: z.string().optional().default(""),
	created_at: z.string(),
	updated_at: z.string(),
	people: z.array(activityPersonSchema).optional().default([]),
})

export const journalListSchema = z.object({
	items: z.array(journalActivitySchema),
	total: z.number(),
	page: z.number(),
	page_size: z.number(),
})

export type ActivityPerson = z.infer<typeof activityPersonSchema>
export type JournalActivity = z.infer<typeof journalActivitySchema>
export type JournalList = z.infer<typeof journalListSchema>

export const journalRequestSchema = z.object({
	title: z.string().min(1),
	content: z.string().optional().default(""),
	occurred_at_date: z.string(),
	occurred_at_time: z.string().optional().default(""),
	person_ids: z.array(z.number()).optional().default([]),
})

export type JournalRequest = z.infer<typeof journalRequestSchema>
