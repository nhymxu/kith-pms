export type ChannelKind = "ntfy" | "telegram" | "discord" | "webhook";

export type ChannelFieldKey =
	| "server_url"
	| "topic"
	| "token"
	| "bot_token"
	| "chat_id"
	| "api_url"
	| "webhook_url"
	| "url"
	| "secret";

export type ChannelField = {
	key: ChannelFieldKey;
	secret?: boolean;
	optional?: boolean;
	type?: "url";
	// Secret the browser mints; shown once with copy and regenerate.
	generate?: boolean;
	placeholder?: string;
};

// Mirrors the Go adapters in internal/notify/channels: field keys, which are
// write-only secrets, and which may be left blank.
export const CHANNEL_KINDS: Record<ChannelKind, { fields: ChannelField[] }> = {
	ntfy: {
		fields: [
			{
				key: "server_url",
				type: "url",
				optional: true,
				placeholder: "https://ntfy.sh",
			},
			{ key: "topic", secret: true },
			{ key: "token", secret: true, optional: true },
		],
	},
	telegram: {
		fields: [
			{ key: "bot_token", secret: true, placeholder: "123456:ABC-DEF..." },
			{ key: "chat_id", placeholder: "123456789" },
			{
				key: "api_url",
				type: "url",
				optional: true,
				placeholder: "https://api.telegram.org",
			},
		],
	},
	discord: {
		fields: [
			{
				key: "webhook_url",
				type: "url",
				secret: true,
				placeholder: "https://discord.com/api/webhooks/...",
			},
		],
	},
	webhook: {
		fields: [
			{
				key: "url",
				type: "url",
				placeholder: "https://example.com/kith-webhook",
			},
			{ key: "secret", secret: true, generate: true },
		],
	},
};

export function generateWebhookSecret(): string {
	const b = crypto.getRandomValues(new Uint8Array(32));
	return Array.from(b, (x) => x.toString(16).padStart(2, "0")).join("");
}

export function isChannelKind(k: string): k is ChannelKind {
	return k in CHANNEL_KINDS;
}
