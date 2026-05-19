import { apiFetch } from "#/lib/api-client";
import { type UserSettings, userSettingsSchema } from "#/schemas/settings";

type Envelope<T> = { data: T };

export async function getSettings(): Promise<UserSettings> {
	const res = await apiFetch<Envelope<unknown>>("/v1/settings");
	return userSettingsSchema.parse(res.data);
}

export async function updateSettings(
	body: UserSettings,
): Promise<UserSettings> {
	const res = await apiFetch<Envelope<unknown>>("/v1/settings", {
		method: "PUT",
		body: JSON.stringify(body),
	});
	return userSettingsSchema.parse(res.data);
}
