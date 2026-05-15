import { z } from "zod"

export const userSchema = z.object({
	id: z.number(),
	created_at: z.string(),
})

export type User = z.infer<typeof userSchema>

export const loginResponseSchema = z.object({
	logged_in: z.boolean(),
})

export const logoutResponseSchema = z.object({
	logged_out: z.boolean(),
})

export const passwordChangeResponseSchema = z.object({
	password_changed: z.boolean(),
})
