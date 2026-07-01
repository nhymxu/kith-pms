// Relationship graph endpoint — read-only, no CSRF header required (GET).
import { apiFetch } from "../lib/api-client";

export interface GraphNode {
	id: number;
	name: string;
	nickname: string;
	avatar: string;
	group: string;
	is_self: boolean;
	date_of_birth?: string | null;
	last_contact_at?: string | null;
}

export interface GraphLink {
	source: number;
	target: number;
	type: string;
	reverse_type: string;
}

export interface RelationshipGraph {
	nodes: GraphNode[];
	links: GraphLink[];
}

type Envelope<T> = { data: T };

export async function getRelationshipGraph(
	personId?: number,
): Promise<RelationshipGraph> {
	const path =
		personId != null
			? `/v1/relationships/graph?person_id=${personId}`
			: "/v1/relationships/graph";
	const res = await apiFetch<Envelope<RelationshipGraph>>(path);
	return res.data;
}

const BULK_RELATIONSHIP_CHUNK = 50;

type BulkRelationship = {
	to_person_id: number;
	relationship_type_id: number;
	notes?: string;
};

async function postBulkBatch(
	personId: number,
	batch: BulkRelationship[],
): Promise<{ created: number; skipped: number }> {
	const res = await apiFetch<Envelope<{ created: number; skipped: number }>>(
		`/v1/people/${personId}/relationships/bulk`,
		{
			method: "POST",
			body: JSON.stringify({ relationships: batch }),
		},
	);
	return res.data;
}

export async function bulkCreateRelationships(
	personId: number,
	relationships: BulkRelationship[],
): Promise<{ created: number; skipped: number }> {
	const acc = { created: 0, skipped: 0 };
	for (let i = 0; i < relationships.length; i += BULK_RELATIONSHIP_CHUNK) {
		const batch = relationships.slice(i, i + BULK_RELATIONSHIP_CHUNK);
		try {
			const r = await postBulkBatch(personId, batch);
			acc.created += r.created;
			acc.skipped += r.skipped;
		} catch (err) {
			(err as Error & { partial?: typeof acc }).partial = { ...acc };
			throw err;
		}
	}
	return acc;
}
