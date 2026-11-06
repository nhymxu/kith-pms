import { z } from "zod";
import { apiFetch } from "#/lib/api-client";

type Envelope<T> = { data: T };

const feedTokenSchema = z.object({ token: z.string().min(1) });

export async function getCalendarFeedToken(): Promise<string> {
	const res = await apiFetch<Envelope<unknown>>("/v1/calendar/feed-token");
	return feedTokenSchema.parse(res.data).token;
}

export async function regenerateCalendarFeedToken(): Promise<string> {
	const res = await apiFetch<Envelope<unknown>>(
		"/v1/calendar/feed-token/regenerate",
		{ method: "POST" },
	);
	return feedTokenSchema.parse(res.data).token;
}
