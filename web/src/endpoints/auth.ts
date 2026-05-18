// Auth endpoints: login, logout, logoutAll, me, changePassword
import { apiFetch } from "../lib/api-client"
import {
	userSchema,
	loginResponseSchema,
	logoutResponseSchema,
	passwordChangeResponseSchema,
	type User,
} from "../schemas/auth"

type Envelope<T> = { data: T }

export async function login(password: string): Promise<boolean> {
	const res = await apiFetch<Envelope<unknown>>("/v1/auth/login", {
		method: "POST",
		body: JSON.stringify({ password }),
		skipSessionLost: true,
	})
	const parsed = loginResponseSchema.parse(res.data)
	return parsed.logged_in
}

export async function logout(): Promise<boolean> {
	const res = await apiFetch<Envelope<unknown>>("/v1/auth/logout", { method: "POST" })
	const parsed = logoutResponseSchema.parse(res.data)
	return parsed.logged_out
}

export async function logoutAll(): Promise<boolean> {
	const res = await apiFetch<Envelope<unknown>>("/v1/auth/logout-all", { method: "POST" })
	const parsed = logoutResponseSchema.parse(res.data)
	return parsed.logged_out
}

export async function me(): Promise<User> {
	const res = await apiFetch<Envelope<unknown>>("/v1/auth/me")
	return userSchema.parse(res.data)
}

export async function changePassword(
	currentPassword: string,
	newPassword: string,
	confirmPassword: string,
): Promise<boolean> {
	const res = await apiFetch<Envelope<unknown>>("/v1/auth/password", {
		method: "POST",
		body: JSON.stringify({ current_password: currentPassword, new_password: newPassword, confirm_password: confirmPassword }),
	})
	const parsed = passwordChangeResponseSchema.parse(res.data)
	return parsed.password_changed
}
