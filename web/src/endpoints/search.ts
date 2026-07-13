import { apiFetch } from "#/lib/api-client";
import { type SearchResult, searchResultSchema } from "#/schemas/search";

type Envelope<T> = { data: T };

export async function searchAll(q: string): Promise<SearchResult> {
	const qs = new URLSearchParams();
	if (q) qs.set("q", q);

	const res = await apiFetch<Envelope<unknown>>(`/v1/search?${qs.toString()}`);
	return searchResultSchema.parse(res.data);
}
