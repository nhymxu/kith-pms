import { z } from "zod"

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
])

export const auditActionSchema = z.enum(["create", "update", "delete"])

// Audit handler returns a manually-constructed map with lowercase json keys
export const auditEntrySchema = z.object({
	id: z.number(),
	entity_type: auditEntityTypeSchema,
	entity_id: z.number(),
	entity_name: z.string(),
	action: auditActionSchema,
	actor_id: z.number().nullable().optional(),
	created_at: z.string(),
})

// Audit list response wraps entries with pagination; handler returns full envelope inline
export const auditListEnvelopeSchema = z.object({
	data: z.array(auditEntrySchema),
	page: z.number(),
	page_size: z.number(),
	has_more: z.boolean(),
})

export type AuditEntityType = z.infer<typeof auditEntityTypeSchema>
export type AuditAction = z.infer<typeof auditActionSchema>
export type AuditEntry = z.infer<typeof auditEntrySchema>
export type AuditListEnvelope = z.infer<typeof auditListEnvelopeSchema>
