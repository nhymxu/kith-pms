import { z } from "zod"

export const labelSchema = z.object({
	id: z.number(),
	name: z.string(),
	color: z.string(),
	created_at: z.string(),
	count: z.number().optional().default(0),
})

export const labelListSchema = z.array(labelSchema)

export type Label = z.infer<typeof labelSchema>
export type LabelList = z.infer<typeof labelListSchema>

export const labelRequestSchema = z.object({
	name: z.string().min(1, "Name is required").max(64),
	color: z.string().regex(/^#[0-9a-fA-F]{6}$/, "Must be a valid hex color e.g. #a1b2c3"),
})

export type LabelRequest = z.infer<typeof labelRequestSchema>
