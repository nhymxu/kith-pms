// Theme registry + runtime. DB is source of truth; localStorage (kith_user_prefs)
// is a write-through cache. The FOUC guard in index.html duplicates THEMES —
// keep the two lists (and the Go validThemes map) in sync.

import { getUserPrefs, saveUserPrefs } from "#/lib/format-datetime";

export const THEMES = [
	"quiet-ink",
	"warm-album",
	"bold-press",
	"nightdesk",
	"softclay",
	"ledger",
] as const;

export type Theme = (typeof THEMES)[number];

export const DEFAULT_THEME: Theme = "quiet-ink";

export interface ThemeMeta {
	id: Theme;
	// Literal preview hex for picker swatch chips (not live theme tokens).
	swatch: { bg: string; panel: string; accent: string; ink: string };
}

export const THEME_META: ThemeMeta[] = [
	{
		id: "quiet-ink",
		swatch: {
			bg: "#ffffff",
			panel: "#ffffff",
			accent: "#4f46e5",
			ink: "#18181b",
		},
	},
	{
		id: "warm-album",
		swatch: {
			bg: "#f7f4ec",
			panel: "#fffdf7",
			accent: "#33604a",
			ink: "#2b2620",
		},
	},
	{
		id: "bold-press",
		swatch: {
			bg: "#f2eefe",
			panel: "#ffffff",
			accent: "#a388ee",
			ink: "#18181b",
		},
	},
	{
		id: "nightdesk",
		swatch: {
			bg: "#12151c",
			panel: "#191e28",
			accent: "#e3ae4f",
			ink: "#e9e7df",
		},
	},
	{
		id: "softclay",
		swatch: {
			bg: "#edeff5",
			panel: "#ffffff",
			accent: "#5b6ee8",
			ink: "#23262f",
		},
	},
	{
		id: "ledger",
		swatch: {
			bg: "#fcfcfb",
			panel: "#ffffff",
			accent: "#0f766e",
			ink: "#161615",
		},
	},
];

function isTheme(value: unknown): value is Theme {
	return (
		typeof value === "string" && (THEMES as readonly string[]).includes(value)
	);
}

// Omit the attribute for the default theme so :root stays authoritative.
export function applyTheme(theme: Theme): void {
	if (theme === DEFAULT_THEME) {
		delete document.documentElement.dataset.theme;
	} else {
		document.documentElement.dataset.theme = theme;
	}
}

export function getTheme(): Theme {
	const t = getUserPrefs().theme;
	return isTheme(t) ? t : DEFAULT_THEME;
}

// Applies + caches. Persisting to the API (PUT /v1/settings) is the caller's job.
export function setTheme(theme: Theme): void {
	applyTheme(theme);
	saveUserPrefs({ theme });
}
