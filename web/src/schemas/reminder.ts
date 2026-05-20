import { z } from "zod";

export const recurrenceRuleSchema = z.object({
	type: z.enum([
		"daily",
		"weekly",
		"monthly",
		"yearly",
		"custom",
		"relative_contact",
		"day_of_week",
	]),
	interval: z.number().optional(),
	unit: z.enum(["days", "weeks", "months"]).optional(),
	day_of_week: z.number().min(0).max(6).optional(),
});

export type RecurrenceRule = z.infer<typeof recurrenceRuleSchema>;

export const reminderSchema = z.object({
	id: z.number(),
	title: z.string(),
	notes: z.string().optional().default(""),
	due_date: z.string(),
	person_id: z.number().nullable().optional(),
	important_date_id: z.number().nullable().optional(),
	completed: z.boolean().optional().default(false),
	completed_at: z.string().nullable().optional(),
	recurrence_rule: recurrenceRuleSchema.nullable().optional(),
	recurrence_end_date: z.string().nullable().optional(),
	created_at: z.string(),
	updated_at: z.string(),
});

export const reminderWithPersonSchema = reminderSchema.extend({
	person_name: z.string().optional().default(""),
});

export const reminderListSchema = z.array(reminderWithPersonSchema);

export type Reminder = z.infer<typeof reminderSchema>;
export type ReminderWithPerson = z.infer<typeof reminderWithPersonSchema>;
export type ReminderList = z.infer<typeof reminderListSchema>;

export const reminderRequestSchema = z.object({
	title: z.string().min(1),
	notes: z.string().optional().default(""),
	due_date: z.string(),
	person_id: z.number().nullable().optional(),
	important_date_id: z.number().nullable().optional(),
	recurrence_rule: recurrenceRuleSchema.nullable().optional(),
	recurrence_end_date: z.string().nullable().optional(),
});

export type ReminderRequest = z.infer<typeof reminderRequestSchema>;
