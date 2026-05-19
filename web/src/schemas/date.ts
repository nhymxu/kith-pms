import { z } from "zod";

export const importantDateSchema = z.object({
	id: z.number(),
	person_id: z.number(),
	kind: z.string(),
	label: z.string().optional().default(""),
	date_value: z.string(),
	recurring: z.boolean().optional().default(false),
	notes: z.string().optional().default(""),
	position: z.number().optional().default(0),
	created_at: z.string(),
});

export const importantDateListSchema = z.array(importantDateSchema);

export const personRefSchema = z.object({
	id: z.number(),
	name: z.string(),
});

export const upcomingDateItemSchema = z.object({
	person: personRefSchema,
	kind: z.string(),
	date_value: z.string(),
	years_since: z.number(),
	next_occurrence: z.string(),
});

export const upcomingDatesListSchema = z.array(upcomingDateItemSchema);

export type ImportantDate = z.infer<typeof importantDateSchema>;
export type ImportantDateList = z.infer<typeof importantDateListSchema>;
export type PersonRef = z.infer<typeof personRefSchema>;
export type UpcomingDateItem = z.infer<typeof upcomingDateItemSchema>;
export type UpcomingDatesList = z.infer<typeof upcomingDatesListSchema>;

export const importantDateRequestSchema = z.object({
	kind: z.string(),
	label: z.string().optional().default(""),
	date_value: z.string(),
	recurring: z.boolean().optional().default(false),
	notes: z.string().optional().default(""),
	position: z.number().optional().default(0),
});

export const datesReplaceRequestSchema = z.object({
	dates: z.array(importantDateRequestSchema),
});

export type ImportantDateRequest = z.infer<typeof importantDateRequestSchema>;
export type DatesReplaceRequest = z.infer<typeof datesReplaceRequestSchema>;
