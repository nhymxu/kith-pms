import { apiFetch } from "#/lib/api-client";
import {
	type SettingsResponse,
	settingsResponseSchema,
	type UserSettings,
} from "#/schemas/settings";

type Envelope<T> = { data: T };

export async function getSettings(): Promise<SettingsResponse> {
	const res = await apiFetch<Envelope<unknown>>("/v1/settings");
	return settingsResponseSchema.parse(res.data);
}

export async function updateSettings(
	body: UserSettings,
): Promise<SettingsResponse> {
	const res = await apiFetch<Envelope<unknown>>("/v1/settings", {
		method: "PUT",
		body: JSON.stringify(body),
	});
	return settingsResponseSchema.parse(res.data);
}
