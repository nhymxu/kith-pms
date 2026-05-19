import { z } from "zod";

export const giftDirectionSchema = z.enum(["given", "received", "planned"]);
export const giftDebtTypeSchema = z.enum(["i_owe", "they_owe", ""]);

export const giftSchema = z.object({
	id: z.number(),
	person_id: z.number(),
	title: z.string(),
	direction: giftDirectionSchema,
	date: z.string().optional().default(""),
	notes: z.string().optional().default(""),
	amount_cents: z.number().nullable().optional(),
	currency: z.string().optional().default("USD"),
	debt_type: giftDebtTypeSchema.optional().default(""),
	image_path: z.string().optional().default(""),
	image_mime_type: z.string().optional().default(""),
	created_at: z.string(),
	updated_at: z.string(),
});

export const giftWithPersonSchema = giftSchema.extend({
	person_name: z.string(),
});

export const giftListSchema = z.object({
	items: z.array(giftWithPersonSchema),
	total: z.number(),
	page: z.number(),
	page_size: z.number(),
});

export type GiftDirection = z.infer<typeof giftDirectionSchema>;
export type GiftDebtType = z.infer<typeof giftDebtTypeSchema>;
export type Gift = z.infer<typeof giftSchema>;
export type GiftWithPerson = z.infer<typeof giftWithPersonSchema>;
export type GiftList = z.infer<typeof giftListSchema>;

export const giftRequestSchema = z.object({
	person_id: z.number(),
	title: z.string().min(1),
	direction: giftDirectionSchema.optional().default("planned"),
	date: z.string().optional().default(""),
	notes: z.string().optional().default(""),
	amount_cents: z.number().nullable().optional(),
	currency: z.string().optional().default("USD"),
	debt_type: giftDebtTypeSchema.optional().default(""),
});

export type GiftRequest = z.infer<typeof giftRequestSchema>;
