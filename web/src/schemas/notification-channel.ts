import { z } from "zod";

export const notificationChannelSchema = z.object({
	id: z.number().int(),
	kind: z.string(),
	name: z.string(),
	enabled: z.boolean(),
	status: z.enum(["ok", "secret_unreadable"]),
	config: z.record(z.string(), z.unknown()),
	created_at: z.string(),
});

export type NotificationChannel = z.infer<typeof notificationChannelSchema>;

export const notificationChannelListSchema = z.object({
	channels: z.array(notificationChannelSchema),
	kinds: z.array(z.string()),
});

export type NotificationChannelList = z.infer<
	typeof notificationChannelListSchema
>;

export const channelTestResultSchema = z.object({
	ok: z.boolean(),
	class: z.string().optional(),
	status: z.number().optional(),
});

export type ChannelTestResult = z.infer<typeof channelTestResultSchema>;

export const webPushKeySchema = z.object({ public_key: z.string() });

export type WebPushKey = z.infer<typeof webPushKeySchema>;
