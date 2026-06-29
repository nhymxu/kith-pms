import { getAvatarUrl } from "#/endpoints/people";

const cache = new Map<number, HTMLImageElement>();

// Returns a cached HTMLImageElement for the given person, or undefined while loading.
// onLoad fires after image loads so the caller can trigger a canvas repaint.
export function getOrLoadImage(
	id: number,
	onLoad: () => void,
): HTMLImageElement | undefined {
	const hit = cache.get(id);
	if (hit) return hit;

	const img = new Image();
	img.onload = () => {
		cache.set(id, img);
		onLoad();
	};
	img.onerror = () => {
		// Keep a sentinel so we don't retry indefinitely.
		cache.set(id, img);
	};
	// Same-origin authed URL — browser sends the session cookie automatically.
	img.src = getAvatarUrl(id);

	return undefined;
}
