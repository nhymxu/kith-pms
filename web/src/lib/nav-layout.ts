// Nav layout registry + runtime. DB is source of truth; localStorage (kith_user_prefs)
// is a write-through cache. Unlike theme, there is no index.html FOUC guard — the
// shell is React-rendered, so the cached value is read at component render init.
// Keep this list in sync with the Go validNavLayout map and the zod enum in
// web/src/schemas/settings.ts.

import { getUserPrefs, saveUserPrefs } from "#/lib/format-datetime";

export const NAV_LAYOUTS = ["top", "side"] as const;

export type NavLayout = (typeof NAV_LAYOUTS)[number];

export const DEFAULT_NAV_LAYOUT: NavLayout = "top";

export interface NavLayoutMeta {
	id: NavLayout;
	label: string;
	description: string;
}

export const NAV_LAYOUT_META: NavLayoutMeta[] = [
	{
		id: "top",
		label: "Top bar",
		description: "Navigation lives in a horizontal bar above the page content.",
	},
	{
		id: "side",
		label: "Side rail",
		description: "Navigation lives in a vertical rail along the left edge.",
	},
];

function isNavLayout(value: unknown): value is NavLayout {
	return (
		typeof value === "string" &&
		(NAV_LAYOUTS as readonly string[]).includes(value)
	);
}

export function getNavLayout(): NavLayout {
	const v = getUserPrefs().navLayout;
	return isNavLayout(v) ? v : DEFAULT_NAV_LAYOUT;
}

// Caches only — no DOM write (unlike applyTheme). Persisting to the API
// (PUT /v1/settings) is the caller's job.
export function setNavLayout(layout: NavLayout): void {
	saveUserPrefs({ navLayout: layout });
}
