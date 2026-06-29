import { z } from "zod";

// Mirrors audit.EntityType and audit.Action Go constants
export const auditEntityTypeSchema = z.enum([
	"person",
	"journal",
	"label",
	"reminder",
	"date",
	"work_history",
	"gift",
	"relationship_type",
	"person_relationship",
]);

export const auditActionSchema = z.enum(["create", "update", "delete"]);

export const auditChangeSchema = z.object({
	field: z.string(),
	old: z.unknown(),
	new: z.unknown(),
});

export const auditMetadataSchema = z.object({
	detail_action: z.string().optional(),
	changes: z.array(auditChangeSchema).optional(),
});

// Audit handler returns a manually-constructed map with lowercase json keys
export const auditEntrySchema = z.object({
	id: z.number(),
	entity_type: auditEntityTypeSchema,
	entity_id: z.number(),
	entity_name: z.string(),
	action: auditActionSchema,
	actor_id: z.number().nullable().optional(),
	metadata: auditMetadataSchema.nullable().optional(),
	created_at: z.string(),
});

// Audit list response wraps entries with pagination; handler returns full envelope inline
export const auditListEnvelopeSchema = z.object({
	data: z.array(auditEntrySchema),
	page: z.number(),
	page_size: z.number(),
	has_more: z.boolean(),
});

export type AuditEntityType = z.infer<typeof auditEntityTypeSchema>;
export type AuditAction = z.infer<typeof auditActionSchema>;
export type AuditChange = z.infer<typeof auditChangeSchema>;
export type AuditMetadata = z.infer<typeof auditMetadataSchema>;
export type AuditEntry = z.infer<typeof auditEntrySchema>;
export type AuditListEnvelope = z.infer<typeof auditListEnvelopeSchema>;
