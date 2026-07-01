import { apiFetch } from "../lib/api-client";
import {
	type PeopleLabel,
	type PeopleLabelList,
	type PeopleLabelRequest,
	peopleLabelListSchema,
	peopleLabelSchema,
} from "../schemas/people-label";

type Envelope<T> = { data: T };

export async function listPeopleLabels(): Promise<PeopleLabelList> {
	const res = await apiFetch<Envelope<unknown>>("/v1/people-labels");
	return peopleLabelListSchema.parse(res.data ?? []);
}

export async function getPeopleLabel(id: number): Promise<PeopleLabel> {
	const res = await apiFetch<Envelope<unknown>>(`/v1/people-labels/${id}`);
	return peopleLabelSchema.parse(res.data);
}

export async function createPeopleLabel(
	body: PeopleLabelRequest,
): Promise<number> {
	const res = await apiFetch<Envelope<{ id: number }>>("/v1/people-labels", {
		method: "POST",
		body: JSON.stringify(body),
	});
	return res.data.id;
}

export async function updatePeopleLabel(
	id: number,
	body: PeopleLabelRequest,
): Promise<void> {
	await apiFetch(`/v1/people-labels/${id}`, {
		method: "PUT",
		body: JSON.stringify(body),
	});
}

export async function deletePeopleLabel(id: number): Promise<void> {
	await apiFetch(`/v1/people-labels/${id}`, { method: "DELETE" });
}

export async function bulkAssignLabel(
	labelId: number,
	personIds: number[],
): Promise<{ attached: number; already_had_label: number }> {
	const res = await apiFetch<
		Envelope<{ attached: number; already_had_label: number }>
	>("/v1/people-labels/bulk-assign", {
		method: "POST",
		body: JSON.stringify({ label_id: labelId, person_ids: personIds }),
	});
	return res.data;
}

export async function previewConnectAll(
	labelId: number,
): Promise<{ member_count: number; pair_count: number }> {
	const res = await apiFetch<
		Envelope<{ member_count: number; pair_count: number }>
	>(`/v1/people-labels/${labelId}/connect-all/preview`);
	return res.data;
}

export async function connectAllMembers(
	labelId: number,
	relationshipTypeId: number,
): Promise<{ created: number; skipped: number; total_members: number }> {
	const res = await apiFetch<
		Envelope<{ created: number; skipped: number; total_members: number }>
	>(`/v1/people-labels/${labelId}/connect-all`, {
		method: "POST",
		body: JSON.stringify({ relationship_type_id: relationshipTypeId }),
	});
	return res.data;
}
