// Relationship-types endpoints: list, create, update, delete
import { apiFetch } from "../lib/api-client";
import {
	type RelationshipType,
	type RelationshipTypeList,
	type RelationshipTypeRequest,
	relationshipTypeListSchema,
	relationshipTypeSchema,
} from "../schemas/relationship-type";

type Envelope<T> = { data: T };

export async function listRelationshipTypes(): Promise<RelationshipTypeList> {
	const res = await apiFetch<Envelope<unknown>>("/v1/relationship-types");
	return relationshipTypeListSchema.parse(res.data);
}

export async function createRelationshipType(
	body: RelationshipTypeRequest,
): Promise<RelationshipType> {
	const res = await apiFetch<Envelope<unknown>>("/v1/relationship-types", {
		method: "POST",
		body: JSON.stringify({ name: body.name, reverse_name: body.reverse_name }),
	});
	return relationshipTypeSchema.parse(res.data);
}

export async function updateRelationshipType(
	id: number,
	body: RelationshipTypeRequest,
): Promise<void> {
	await apiFetch(`/v1/relationship-types/${id}`, {
		method: "PUT",
		body: JSON.stringify({ name: body.name, reverse_name: body.reverse_name }),
	});
}

export async function deleteRelationshipType(id: number): Promise<void> {
	await apiFetch(`/v1/relationship-types/${id}`, { method: "DELETE" });
}
