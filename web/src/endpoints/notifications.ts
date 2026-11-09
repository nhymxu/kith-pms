import { apiFetch } from "#/lib/api-client";
import {
	type ChannelTestResult,
	channelTestResultSchema,
	type NotificationChannel,
	type NotificationChannelList,
	notificationChannelListSchema,
	notificationChannelSchema,
} from "#/schemas/notification-channel";

type Envelope<T> = { data: T };

export async function listNotificationChannels(): Promise<NotificationChannelList> {
	const res = await apiFetch<Envelope<unknown>>("/v1/notifications/channels");
	return notificationChannelListSchema.parse(res.data);
}

export async function setNotificationChannelEnabled(
	id: number,
	enabled: boolean,
): Promise<NotificationChannel> {
	const res = await apiFetch<Envelope<unknown>>(
		`/v1/notifications/channels/${id}`,
		{ method: "PATCH", body: JSON.stringify({ enabled }) },
	);
	return notificationChannelSchema.parse(res.data);
}

export async function deleteNotificationChannel(id: number): Promise<void> {
	await apiFetch(`/v1/notifications/channels/${id}`, { method: "DELETE" });
}

export async function testNotificationChannel(
	id: number,
): Promise<ChannelTestResult> {
	const res = await apiFetch<Envelope<unknown>>(
		`/v1/notifications/channels/${id}/test`,
		{ method: "POST" },
	);
	return channelTestResultSchema.parse(res.data);
}

export async function createNotificationChannel(input: {
	kind: string;
	name: string;
	config: Record<string, string>;
}): Promise<NotificationChannel> {
	const res = await apiFetch<Envelope<unknown>>("/v1/notifications/channels", {
		method: "POST",
		body: JSON.stringify(input),
	});
	return notificationChannelSchema.parse(res.data);
}

// Secrets left out of config keep their stored value.
export async function updateNotificationChannelConfig(
	id: number,
	config: Record<string, string>,
): Promise<NotificationChannel> {
	const res = await apiFetch<Envelope<unknown>>(
		`/v1/notifications/channels/${id}`,
		{ method: "PUT", body: JSON.stringify(config) },
	);
	return notificationChannelSchema.parse(res.data);
}
