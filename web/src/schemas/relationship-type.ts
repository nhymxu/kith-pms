import { z } from "zod";

export const relationshipTypeSchema = z.object({
	id: z.number(),
	name: z.string(),
	reverse_name: z.string().optional().default(""),
	inverse_type_id: z.number().nullable().optional(),
	created_at: z.string(),
	usage_count: z.number().optional().default(0),
});

export const relationshipTypeListSchema = z.array(relationshipTypeSchema);

export const relationshipViewSchema = z.object({
	id: z.number(),
	other_person_id: z.number(),
	other_person_name: z.string(),
	other_person_avatar: z.string().optional().default(""),
	type_name: z.string(),
	notes: z.string().optional().default(""),
});

export const relationshipViewListSchema = z.array(relationshipViewSchema);

export type RelationshipType = z.infer<typeof relationshipTypeSchema>;
export type RelationshipTypeList = z.infer<typeof relationshipTypeListSchema>;
export type RelationshipView = z.infer<typeof relationshipViewSchema>;
export type RelationshipViewList = z.infer<typeof relationshipViewListSchema>;

export const relationshipTypeRequestSchema = z.object({
	name: z.string().min(1, "Name is required").max(80),
	reverse_name: z.string().optional().default(""),
});

export const attachRelationshipRequestSchema = z.object({
	to_person_id: z.number().positive(),
	relationship_type_id: z.number().positive(),
	notes: z.string().optional().default(""),
});

export type RelationshipTypeRequest = z.infer<
	typeof relationshipTypeRequestSchema
>;
export type AttachRelationshipRequest = z.infer<
	typeof attachRelationshipRequestSchema
>;
