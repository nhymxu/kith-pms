import { z } from "zod";

export const contactInfoSchema = z.object({
	id: z.number(),
	person_id: z.number(),
	type: z.string(),
	value: z.string(),
	label: z.string(),
	position: z.number(),
});

export const locationSchema = z.object({
	id: z.number(),
	person_id: z.number(),
	type: z.string(),
	address: z.string(),
	city: z.string(),
	country: z.string(),
	postal_code: z.string(),
	position: z.number(),
});

export const labelRefSchema = z.object({
	id: z.number(),
	name: z.string(),
	color: z.string(),
});

export const genderOptions = [
	{ value: "male", label: "Male" },
	{ value: "female", label: "Female" },
	{ value: "rather_not_say", label: "Rather not say" },
] as const;

export type GenderValue = "" | "male" | "female" | "rather_not_say";

export const personSchema = z.object({
	id: z.number(),
	is_self: z.boolean().optional().default(false),
	has_birthday_reminder: z.boolean().optional().default(false),
	prefix: z.string().optional().default(""),
	name: z.string(),
	nickname: z.string().optional().default(""),
	gender: z.string().optional().default(""),
	date_of_birth: z.string().nullable().optional(),
	relationship_type: z.string().optional().default(""),
	other_notes: z.string().optional().default(""),
	avatar_path: z.string().optional().default(""),
	avatar_size: z.number().optional().default(0),
	avatar_uploaded_at: z.string().nullable().optional(),
	last_contact_at: z.string().nullable().optional(),
	created_at: z.string(),
	updated_at: z.string(),
	contacts: z
		.array(contactInfoSchema)
		.nullable()
		.optional()
		.transform((v) => v ?? []),
	locations: z
		.array(locationSchema)
		.nullable()
		.optional()
		.transform((v) => v ?? []),
	labels: z
		.array(labelRefSchema)
		.nullable()
		.optional()
		.transform((v) => v ?? []),
});

export const personListSchema = z.object({
	items: z.array(personSchema),
	total: z.number(),
	page: z.number(),
	page_size: z.number(),
});

export type ContactInfo = z.infer<typeof contactInfoSchema>;
export type Location = z.infer<typeof locationSchema>;
export type LabelRef = z.infer<typeof labelRefSchema>;
export type Person = z.infer<typeof personSchema>;
export type PersonList = z.infer<typeof personListSchema>;

// Request shapes
export const personRequestSchema = z.object({
	name: z.string().min(1),
	nickname: z.string().optional().default(""),
	gender: z.string().optional().default(""),
	relationship_type: z.string().optional().default(""),
	date_of_birth: z.string().optional().default(""),
	create_birthday_reminder: z.boolean().optional().default(false),
	last_contact_at: z.string().nullable().optional().default(""),
	other_notes: z.string().optional().default(""),
	contacts: z
		.array(
			z.object({
				type: z.string(),
				value: z.string(),
				label: z.string(),
				position: z.number(),
			}),
		)
		.optional()
		.default([]),
	locations: z
		.array(
			z.object({
				type: z.string(),
				address: z.string(),
				city: z.string(),
				country: z.string(),
				postal_code: z.string(),
				position: z.number(),
			}),
		)
		.optional()
		.default([]),
});

export type PersonRequest = z.infer<typeof personRequestSchema>;
