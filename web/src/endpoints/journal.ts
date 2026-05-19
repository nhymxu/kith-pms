// Journal endpoints: list, get, create, update, delete
import { apiFetch } from "../lib/api-client";
import {
	type JournalActivity,
	type JournalList,
	type JournalRequest,
	journalActivitySchema,
	journalListSchema,
} from "../schemas/journal";

type Envelope<T> = { data: T };

export interface JournalListParams {
	q?: string;
	page?: number;
	page_size?: number;
	from_date?: string;
	to_date?: string;
	person_ids?: number[];
}

export async function listJournal(
	params: JournalListParams = {},
): Promise<JournalList> {
	const qs = new URLSearchParams();
	if (params.q) qs.set("q", params.q);
	if (params.page) qs.set("page", String(params.page));
	if (params.page_size) qs.set("page_size", String(params.page_size));
	if (params.from_date) qs.set("from_date", params.from_date);
	if (params.to_date) qs.set("to_date", params.to_date);
	if (params.person_ids?.length)
		qs.set("person_ids", params.person_ids.join(","));

	const query = qs.toString();
	const res = await apiFetch<Envelope<unknown>>(
		`/v1/journal${query ? `?${query}` : ""}`,
	);
	return journalListSchema.parse(res.data);
}

export async function getJournalEntry(id: number): Promise<JournalActivity> {
	const res = await apiFetch<Envelope<unknown>>(`/v1/journal/${id}`);
	return journalActivitySchema.parse(res.data);
}

export async function createJournalEntry(
	body: JournalRequest,
): Promise<number> {
	const res = await apiFetch<Envelope<{ id: number }>>("/v1/journal", {
		method: "POST",
		body: JSON.stringify(body),
	});
	return res.data.id;
}

export async function updateJournalEntry(
	id: number,
	body: JournalRequest,
): Promise<void> {
	await apiFetch(`/v1/journal/${id}`, {
		method: "PUT",
		body: JSON.stringify(body),
	});
}

export async function deleteJournalEntry(id: number): Promise<void> {
	await apiFetch(`/v1/journal/${id}`, { method: "DELETE" });
}
