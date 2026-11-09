import {
	createNotificationChannel,
	deleteNotificationChannel,
	getWebPushKey,
} from "#/endpoints/notifications";

const CHANNEL_ID_KEY = "kith.webpush.channelId";

export function isPushSupported(): boolean {
	return (
		typeof window !== "undefined" &&
		"serviceWorker" in navigator &&
		"PushManager" in window &&
		"Notification" in window
	);
}

export function urlBase64ToUint8Array(b64: string): Uint8Array<ArrayBuffer> {
	const padded = (b64 + "=".repeat((4 - (b64.length % 4)) % 4))
		.replace(/-/g, "+")
		.replace(/_/g, "/");
	const raw = atob(padded);
	return Uint8Array.from(raw, (c) => c.charCodeAt(0));
}

type Owned = { id: number; createdAt: string };

// SQLite reuses channel ids, so the id alone could point at another browser's
// (or another kind's) channel; created_at pins it to the row this browser made.
export function storedChannel(): Owned | null {
	try {
		const v = JSON.parse(localStorage.getItem(CHANNEL_ID_KEY) ?? "null");
		return Number.isInteger(v?.id) && typeof v?.createdAt === "string"
			? v
			: null;
	} catch {
		return null;
	}
}

export function ownedChannelId(
	channels: { id: number; kind: string; created_at: string }[],
): number | null {
	const mine = storedChannel();
	const hit = channels.find(
		(c) =>
			c.kind === "webpush" &&
			c.id === mine?.id &&
			c.created_at === mine?.createdAt,
	);
	return hit ? hit.id : null;
}

async function currentSubscription() {
	const reg = await navigator.serviceWorker.register("/sw.js");
	await navigator.serviceWorker.ready;
	return { reg, sub: await reg.pushManager.getSubscription() };
}

// A leftover subscription made with an older VAPID key makes subscribe() throw,
// so it is always dropped first.
export async function subscribeThisBrowser(name: string): Promise<number> {
	if ((await Notification.requestPermission()) !== "granted") {
		throw new Error("permission_denied");
	}
	const { public_key: publicKey } = await getWebPushKey();
	const { reg, sub: old } = await currentSubscription();
	await old?.unsubscribe();

	const sub = await reg.pushManager.subscribe({
		userVisibleOnly: true,
		applicationServerKey: urlBase64ToUint8Array(publicKey),
	});
	const json = sub.toJSON();
	try {
		const channel = await createNotificationChannel({
			kind: "webpush",
			name,
			config: {
				endpoint: sub.endpoint,
				p256dh: json.keys?.p256dh ?? "",
				auth: json.keys?.auth ?? "",
			},
		});
		localStorage.setItem(
			CHANNEL_ID_KEY,
			JSON.stringify({ id: channel.id, createdAt: channel.created_at }),
		);
		return channel.id;
	} catch (err) {
		await sub.unsubscribe();
		throw err;
	}
}

export async function unsubscribeThisBrowser(channelId: number | null) {
	if (channelId !== null) {
		await deleteNotificationChannel(channelId);
	}
	localStorage.removeItem(CHANNEL_ID_KEY);
	const { sub } = await currentSubscription();
	await sub?.unsubscribe();
}
