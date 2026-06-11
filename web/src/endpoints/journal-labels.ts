import { apiFetch } from "#/lib/api-client";
import {
	type JournalLabel,
	type JournalLabelRequest,
	journalLabelListSchema,
	journalLabelSchema,
} from "#/schemas/journal-label";

type Envelope<T> = { data: T };

export async function listJournalLabels(): Promise<JournalLabel[]> {
	const res = await apiFetch<Envelope<unknown>>("/v1/journal-labels");
	return journalLabelListSchema.parse(res.data);
}

export async function getJournalLabel(id: number): Promise<JournalLabel> {
	const res = await apiFetch<Envelope<unknown>>(`/v1/journal-labels/${id}`);
	return journalLabelSchema.parse(res.data);
}

export async function createJournalLabel(
	body: JournalLabelRequest,
): Promise<number> {
	const res = await apiFetch<Envelope<{ id: number }>>("/v1/journal-labels", {
		method: "POST",
		body: JSON.stringify(body),
	});
	return res.data.id;
}

export async function updateJournalLabel(
	id: number,
	body: JournalLabelRequest,
): Promise<void> {
	await apiFetch(`/v1/journal-labels/${id}`, {
		method: "PUT",
		body: JSON.stringify(body),
	});
}

export async function deleteJournalLabel(id: number): Promise<void> {
	await apiFetch(`/v1/journal-labels/${id}`, { method: "DELETE" });
}
