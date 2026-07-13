// Reads a CSS custom property (design token) off the document root.
// This is the one legitimate place JS reads theme tokens directly —
// canvas/SVG paint code cannot consume Tailwind classes, so it needs the
// resolved value at draw time. Everything else in the app must use the
// semantic utility classes instead (see styles.css).
export function readToken(name: string, fallback: string): string {
	if (typeof document === "undefined") return fallback;
	const value = getComputedStyle(document.documentElement)
		.getPropertyValue(name)
		.trim();
	return value || fallback;
}
