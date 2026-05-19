// People endpoints: CRUD + avatar upload/delete + label attach/detach + relationships
import { apiFetch } from "../lib/api-client";
import {
	type Person,
	type PersonList,
	type PersonRequest,
	personListSchema,
	personSchema,
} from "../schemas/person";
import {
	type AttachRelationshipRequest,
	type RelationshipViewList,
	relationshipViewListSchema,
} from "../schemas/relationship-type";
import { type WorkEntry, workEntryListSchema } from "../schemas/work-history";

type Envelope<T> = { data: T };

export interface PeopleListParams {
	q?: string;
	page?: number;
	page_size?: number;
	labels?: number[];
}

export async function listPeople(
	params: PeopleListParams = {},
): Promise<PersonList> {
	const qs = new URLSearchParams();
	if (params.q) qs.set("q", params.q);
	if (params.page) qs.set("page", String(params.page));
	if (params.page_size) qs.set("page_size", String(params.page_size));
	if (params.labels?.length) qs.set("labels", params.labels.join(","));

	const query = qs.toString();
	const res = await apiFetch<Envelope<unknown>>(
		`/v1/people${query ? `?${query}` : ""}`,
	);
	return personListSchema.parse(res.data);
}

export async function getPerson(id: number): Promise<Person> {
	const res = await apiFetch<Envelope<unknown>>(`/v1/people/${id}`);
	return personSchema.parse(res.data);
}

export async function createPerson(body: PersonRequest): Promise<number> {
	const res = await apiFetch<Envelope<{ id: number }>>("/v1/people", {
		method: "POST",
		body: JSON.stringify(body),
	});
	return res.data.id;
}

export async function updatePerson(
	id: number,
	body: PersonRequest,
): Promise<void> {
	await apiFetch(`/v1/people/${id}`, {
		method: "PUT",
		body: JSON.stringify(body),
	});
}

export async function deletePerson(id: number): Promise<void> {
	await apiFetch(`/v1/people/${id}`, { method: "DELETE" });
}

// Avatar endpoints

export async function uploadAvatar(
	personId: number,
	file: File,
): Promise<void> {
	const form = new FormData();
	form.append("avatar", file);
	await apiFetch(`/v1/people/${personId}/avatar`, {
		method: "POST",
		body: form,
	});
}

export async function deleteAvatar(personId: number): Promise<void> {
	await apiFetch(`/v1/people/${personId}/avatar`, { method: "DELETE" });
}

// Returns the URL to stream the avatar binary — not a fetch call, used as <img src>.
// Prefixes VITE_API_BASE_URL so dev cross-origin setups work correctly.
export function getAvatarUrl(personId: number): string {
	const base = (import.meta.env.VITE_API_BASE_URL as string | undefined) ?? "";
	return `${base}/v1/people/${personId}/avatar`;
}

// Relationship endpoints

export async function listRelationships(
	personId: number,
): Promise<RelationshipViewList> {
	const res = await apiFetch<Envelope<unknown>>(
		`/v1/people/${personId}/relationships`,
	);
	return relationshipViewListSchema.parse(res.data);
}

export async function attachRelationship(
	personId: number,
	body: AttachRelationshipRequest,
): Promise<number> {
	const res = await apiFetch<Envelope<{ id: number }>>(
		`/v1/people/${personId}/relationships`,
		{
			method: "POST",
			body: JSON.stringify(body),
		},
	);
	return res.data.id;
}

export async function detachRelationship(
	personId: number,
	relId: number,
): Promise<void> {
	await apiFetch(`/v1/people/${personId}/relationships/${relId}`, {
		method: "DELETE",
	});
}

// Work history

export async function listWorkHistory(personId: number): Promise<WorkEntry[]> {
	const res = await apiFetch<{ data: unknown }>(
		`/v1/people/${personId}/work-history`,
	);
	return workEntryListSchema.parse(res.data);
}

export interface WorkEntryRequest {
	company: string;
	title: string;
	start_date: string;
	end_date: string;
	location: string;
	description: string;
	position: number;
}

export async function replaceWorkHistory(
	personId: number,
	body: { entries: WorkEntryRequest[] },
): Promise<void> {
	await apiFetch(`/v1/people/${personId}/work-history`, {
		method: "PUT",
		body: JSON.stringify(body),
	});
}

// Label attach/detach

export async function attachLabel(
	personId: number,
	labelId: number,
): Promise<void> {
	await apiFetch(`/v1/people/${personId}/labels`, {
		method: "POST",
		body: JSON.stringify({ label_id: labelId }),
	});
}

export async function detachLabel(
	personId: number,
	labelId: number,
): Promise<void> {
	await apiFetch(`/v1/people/${personId}/labels/${labelId}`, {
		method: "DELETE",
	});
}
