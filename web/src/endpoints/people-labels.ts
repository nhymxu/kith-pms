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
	return peopleLabelListSchema.parse(res.data);
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
