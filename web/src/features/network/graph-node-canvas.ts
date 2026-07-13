// Soft overlap: pending plan 260629-1555-relationship-network-visualization also
// touches features/network/*. Re-audit for reintroduced palette classes/literal
// hex colors when that plan lands (whichever lands second grep-checks the other).

import { formatPersonName } from "#/lib/format-person-name";
import { readToken } from "#/lib/read-css-token";
import { getOrLoadImage } from "./graph-image-cache";
import type { ColorBy, GraphNode } from "./graph-types";

const DIM_ALPHA = 0.15;

// Canvas paint colors sourced from CSS design tokens (getComputedStyle), the
// one legitimate JS token read in the app — canvas can't consume Tailwind
// classes. Cached here and refreshed on theme change via refreshPaintTokens().
interface PaintTokens {
	accent: string;
	edgeNeutral: string;
	mutedFg: string;
	ink: string;
	dimFg: string;
	edgePalette: string[];
}

function readPaintTokens(): PaintTokens {
	return {
		accent: readToken("--accent", "#4f46e5"),
		edgeNeutral: readToken("--line", "#e4e4e7"),
		mutedFg: readToken("--sub", "#71717a"),
		ink: readToken("--ink", "#18181b"),
		dimFg: readToken("--sub", "#a1a1aa"),
		edgePalette: [
			readToken("--chart-1", "#4f46e5"),
			readToken("--chart-2", "#a1a1aa"),
			readToken("--chart-3", "#14b8a6"),
			readToken("--chart-4", "#f59e0b"),
		],
	};
}

let tokens: PaintTokens = readPaintTokens();

/** Re-reads all paint tokens from the document root. Call after a theme change. */
export function refreshPaintTokens(): void {
	tokens = readPaintTokens();
}

function hexToRgb(hex: string): [number, number, number] | null {
	const m = /^#?([0-9a-f]{6})$/i.exec(hex.trim());
	if (!m) return null;
	const n = parseInt(m[1], 16);
	return [(n >> 16) & 255, (n >> 8) & 255, n & 255];
}

/** Neutral edge color at reduced alpha, for dimmed/inactive links. */
export function dimmedLinkColor(alpha = 0.15): string {
	const rgb = hexToRgb(tokens.edgeNeutral);
	return rgb
		? `rgba(${rgb[0]},${rgb[1]},${rgb[2]},${alpha})`
		: `rgba(228,228,231,${alpha})`;
}

function hashColor(str: string): string {
	let h = 0;
	for (let i = 0; i < str.length; i++) h = (h * 31 + str.charCodeAt(i)) >>> 0;
	return tokens.edgePalette[h % tokens.edgePalette.length] ?? tokens.accent;
}

export function typeColor(type: string): string {
	return hashColor(type);
}

export function labelColor(
	group: string,
	groupColorMap: Map<string, string>,
): string {
	if (!group) return tokens.mutedFg;
	return groupColorMap.get(group) ?? tokens.mutedFg;
}

export function linkColor(type: string, colorBy: ColorBy): string {
	return colorBy === "type" ? typeColor(type) : tokens.edgeNeutral;
}

interface DrawNodeOpts {
	colorBy: ColorBy;
	dimmedNodeIds: Set<number>;
	groupColorMap: Map<string, string>;
	showAvatar: boolean;
	onImageLoad: () => void;
}

export function drawNode(
	node: GraphNode & {
		x?: number;
		y?: number;
		fx?: number | null;
		fy?: number | null;
	},
	ctx: CanvasRenderingContext2D,
	globalScale: number,
	opts: DrawNodeOpts,
): void {
	const { colorBy, dimmedNodeIds, groupColorMap, showAvatar, onImageLoad } =
		opts;
	const x = node.x ?? 0;
	const y = node.y ?? 0;
	const r = node.is_self ? 10 : 7;
	const ringWidth = node.is_self ? 2.5 : 1.5;

	const dimmed = dimmedNodeIds.has(node.id);
	const alpha = dimmed ? DIM_ALPHA : 1;

	ctx.save();
	ctx.globalAlpha = alpha;

	const fill =
		colorBy === "labels"
			? labelColor(node.group, groupColorMap)
			: tokens.mutedFg;

	const img =
		showAvatar && node.avatar
			? getOrLoadImage(node.id, onImageLoad)
			: undefined;
	const avatarReady = img?.complete && img.naturalWidth > 0;

	if (avatarReady && img) {
		// Circular-clipped avatar image
		ctx.save();
		ctx.beginPath();
		ctx.arc(x, y, r, 0, Math.PI * 2);
		ctx.clip();
		ctx.drawImage(img, x - r, y - r, r * 2, r * 2);
		ctx.restore();
	} else {
		// Colored disc
		ctx.beginPath();
		ctx.arc(x, y, r, 0, Math.PI * 2);
		ctx.fillStyle = fill;
		ctx.fill();

		// Initial letter centered in the disc
		const initial = (node.name?.[0] ?? "?").toUpperCase();
		const fontSize = Math.max(r * 1.1, 4);
		ctx.font = `600 ${fontSize}px Inter, sans-serif`;
		// Fixed white — contrasts against arbitrary user-chosen label colors and
		// the 4-slot chart palette, not a themed surface fill.
		ctx.fillStyle = "#ffffff";
		ctx.textAlign = "center";
		ctx.textBaseline = "middle";
		ctx.fillText(initial, x, y);
	}

	// Group/self ring
	ctx.beginPath();
	ctx.arc(x, y, r, 0, Math.PI * 2);
	ctx.strokeStyle = node.is_self ? tokens.accent : fill;
	ctx.lineWidth = ringWidth;
	ctx.stroke();

	// Name label below node
	const fontSize = Math.max(10 / globalScale, 2);
	ctx.font = `${fontSize}px Inter, sans-serif`;
	ctx.fillStyle = dimmed ? tokens.dimFg : tokens.ink;
	ctx.textAlign = "center";
	ctx.textBaseline = "top";
	ctx.fillText(
		formatPersonName(node.name, node.nickname),
		x,
		y + r + 2 / globalScale,
	);

	ctx.restore();
}

export function drawHitArea(
	node: GraphNode & { x?: number; y?: number },
	ctx: CanvasRenderingContext2D,
	globalScale: number,
	color: string,
): void {
	const x = node.x ?? 0;
	const y = node.y ?? 0;
	const r = node.is_self ? 10 : 7;
	ctx.beginPath();
	ctx.arc(x, y, r, 0, Math.PI * 2);
	ctx.fillStyle = color;
	ctx.fill();
	void globalScale;
}

export { DIM_ALPHA };
