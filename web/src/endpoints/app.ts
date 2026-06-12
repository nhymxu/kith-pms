import { z } from "zod";
import { apiFetch } from "#/lib/api-client";

const appInfoSchema = z.object({
	version: z.string(),
	commit: z.string(),
});

export type AppInfo = z.infer<typeof appInfoSchema>;

type Envelope<T> = { data: T };

export async function getAppInfo(): Promise<AppInfo> {
	const res = await apiFetch<Envelope<unknown>>("/v1/app/info");
	return appInfoSchema.parse(res.data);
}
