import { formatPersonName } from "#/lib/format-person-name";
import { getOrLoadImage } from "./graph-image-cache";
import type { ColorBy, GraphNode } from "./graph-types";

const PRIMARY = "#4f46e5";
const NEUTRAL_EDGE = "#e4e4e7";
const MUTED_FG = "#71717a";
const DIM_ALPHA = 0.15;

const EDGE_PALETTE = [
	"#6366f1",
	"#0ea5e9",
	"#10b981",
	"#f59e0b",
	"#ef4444",
	"#8b5cf6",
	"#ec4899",
	"#14b8a6",
];

function hashColor(str: string): string {
	let h = 0;
	for (let i = 0; i < str.length; i++) h = (h * 31 + str.charCodeAt(i)) >>> 0;
	return EDGE_PALETTE[h % EDGE_PALETTE.length] ?? PRIMARY;
}

export function typeColor(type: string): string {
	return hashColor(type);
}

export function labelColor(
	group: string,
	groupColorMap: Map<string, string>,
): string {
	if (!group) return MUTED_FG;
	return groupColorMap.get(group) ?? MUTED_FG;
}

export function linkColor(type: string, colorBy: ColorBy): string {
	return colorBy === "type" ? typeColor(type) : NEUTRAL_EDGE;
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
		colorBy === "labels" ? labelColor(node.group, groupColorMap) : MUTED_FG;

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
		ctx.fillStyle = "#ffffff";
		ctx.textAlign = "center";
		ctx.textBaseline = "middle";
		ctx.fillText(initial, x, y);
	}

	// Group/self ring
	ctx.beginPath();
	ctx.arc(x, y, r, 0, Math.PI * 2);
	ctx.strokeStyle = node.is_self ? PRIMARY : fill;
	ctx.lineWidth = ringWidth;
	ctx.stroke();

	// Name label below node
	const fontSize = Math.max(10 / globalScale, 2);
	ctx.font = `${fontSize}px Inter, sans-serif`;
	ctx.fillStyle = dimmed ? "#a1a1aa" : "#18181b";
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

export { DIM_ALPHA, NEUTRAL_EDGE };
