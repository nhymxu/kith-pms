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

export async function bulkCreateRelationships(
	personId: number,
	relationships: Array<{
		to_person_id: number;
		relationship_type_id: number;
		notes?: string;
	}>,
): Promise<{ created: number; skipped: number }> {
	const res = await apiFetch<Envelope<{ created: number; skipped: number }>>(
		`/v1/people/${personId}/relationships/bulk`,
		{
			method: "POST",
			body: JSON.stringify({ relationships }),
		},
	);
	return res.data;
}
