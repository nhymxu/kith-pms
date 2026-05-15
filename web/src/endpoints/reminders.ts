// Reminders endpoints: list, get, create, update, delete, complete
import { apiFetch } from "../lib/api-client"
import {
	reminderWithPersonSchema,
	reminderListSchema,
	type ReminderWithPerson,
	type ReminderList,
	type ReminderRequest,
} from "../schemas/reminder"

type Envelope<T> = { data: T }

export type ReminderStatus = "upcoming" | "overdue" | "all"

export interface ReminderListParams {
	status?: ReminderStatus
	days?: number
}

export async function listReminders(params: ReminderListParams = {}): Promise<ReminderList> {
	const qs = new URLSearchParams()
	if (params.status && params.status !== "all") qs.set("status", params.status)
	if (params.days) qs.set("days", String(params.days))

	const query = qs.toString()
	const res = await apiFetch<Envelope<unknown>>(`/v1/reminders${query ? `?${query}` : ""}`)
	return reminderListSchema.parse(res.data)
}

export async function getReminder(id: number): Promise<ReminderWithPerson> {
	const res = await apiFetch<Envelope<unknown>>(`/v1/reminders/${id}`)
	return reminderWithPersonSchema.parse(res.data)
}

export async function createReminder(body: ReminderRequest): Promise<number> {
	const res = await apiFetch<Envelope<{ id: number }>>("/v1/reminders", {
		method: "POST",
		body: JSON.stringify(body),
	})
	return res.data.id
}

export async function updateReminder(id: number, body: ReminderRequest): Promise<void> {
	await apiFetch(`/v1/reminders/${id}`, {
		method: "PUT",
		body: JSON.stringify(body),
	})
}

export async function deleteReminder(id: number): Promise<void> {
	await apiFetch(`/v1/reminders/${id}`, { method: "DELETE" })
}

export async function completeReminder(id: number): Promise<void> {
	await apiFetch(`/v1/reminders/${id}/complete`, { method: "PATCH" })
}
