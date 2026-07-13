import { z } from "zod";

export const searchItemSchema = z.object({
	id: z.number(),
	title: z.string(),
	subtitle: z.string(),
	url: z.string(),
});

export const searchResultSchema = z.object({
	people: z.array(searchItemSchema),
	journal: z.array(searchItemSchema),
	gifts: z.array(searchItemSchema),
});

export type SearchItem = z.infer<typeof searchItemSchema>;
export type SearchResult = z.infer<typeof searchResultSchema>;
