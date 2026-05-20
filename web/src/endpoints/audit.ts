// Audit endpoint: list entries with optional entity filters
// Note: audit handler returns its own envelope shape (not the standard {data:...} wrapper)
import { apiFetch } from "../lib/api-client";
import {
	type AuditEntityType,
	type AuditListEnvelope,
	auditListEnvelopeSchema,
} from "../schemas/audit";

export interface AuditListParams {
	entity_type?: AuditEntityType;
	entity_id?: number;
	page?: number;
	from_date?: string;
	to_date?: string;
}

export async function listAudit(
	params: AuditListParams = {},
): Promise<AuditListEnvelope> {
	const qs = new URLSearchParams();
	if (params.entity_type) qs.set("entity_type", params.entity_type);
	if (params.entity_id) qs.set("entity_id", String(params.entity_id));
	if (params.page) qs.set("page", String(params.page));
	if (params.from_date) qs.set("from_date", params.from_date);
	if (params.to_date) qs.set("to_date", params.to_date);

	const query = qs.toString();
	// Audit handler returns its own JSON shape directly (not {data:...}), so parse the raw response
	const raw = await apiFetch<unknown>(`/v1/audit${query ? `?${query}` : ""}`);
	return auditListEnvelopeSchema.parse(raw);
}
