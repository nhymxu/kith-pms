import { z } from "zod"

export const workEntrySchema = z.object({
	id: z.number(),
	person_id: z.number(),
	company: z.string(),
	title: z.string().optional().default(""),
	start_date: z.string(),
	end_date: z.string().optional().default(""),
	location: z.string().optional().default(""),
	description: z.string().optional().default(""),
	position: z.number().optional().default(0),
	created_at: z.string(),
})

export const workEntryListSchema = z.array(workEntrySchema)
export type WorkEntry = z.infer<typeof workEntrySchema>
