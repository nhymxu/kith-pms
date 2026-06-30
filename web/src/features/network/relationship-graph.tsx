import { useSuspenseQuery } from "@tanstack/react-query";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import ForceGraph2D, { type ForceGraphMethods } from "react-force-graph-2d";
import { listPeopleLabels } from "#/endpoints/people-labels";
import { cloneGraphData } from "#/lib/graph-data";
import { keys } from "#/query-keys";
import { GraphControls } from "./graph-controls";
import { GraphLegend } from "./graph-legend";
import { drawHitArea, drawNode, linkColor } from "./graph-node-canvas";
import { GraphNodeProfileCard } from "./graph-node-profile-card";
import type { SelectedNodeInfo } from "./graph-selected-panel";
import type {
	ColorBy,
	GraphLink,
	GraphNode,
	RelationshipGraph as RelationshipGraphData,
} from "./graph-types";

const DOTTED_BG =
	"radial-gradient(circle at 1px 1px,#e4e4e7 1px,transparent 0) 0 0/22px 22px,#fafafa";
const PRIMARY = "#4f46e5";

interface RelationshipGraphProps {
	data: RelationshipGraphData;
	focusNodeId?: number;
	height?: number;
	title?: string;
	showOnlyMine?: boolean;
	onShowOnlyMineChange?: (v: boolean) => void;
	onNodeSelect?: (info: SelectedNodeInfo | null) => void;
	sidebarCollapsed?: boolean;
}

export default function RelationshipGraph({
	data,
	focusNodeId,
	height = 500,
	title,
	showOnlyMine,
	onShowOnlyMineChange,
	onNodeSelect,
	sidebarCollapsed,
}: RelationshipGraphProps) {
	const fgRef = useRef<ForceGraphMethods | undefined>(undefined);
	const containerRef = useRef<HTMLDivElement>(null);
	const [width, setWidth] = useState(600);
	const [colorBy, setColorBy] = useState<ColorBy>("labels");
	const [activeGroup, setActiveGroup] = useState<string | null>(null);
	const [activeRelType, setActiveRelType] = useState<string | null>(null);
	const [showAvatar, setShowAvatar] = useState(false);
	const [repaintKey, setRepaintKey] = useState(0);
	const [profileCard, setProfileCard] = useState<{
		node: GraphNode;
		posX: number;
		posY: number;
	} | null>(null);
	const [tooltip, setTooltip] = useState<{
		name: string;
		group: string;
		isSelf: boolean;
		relTypes: string[];
		x: number;
		y: number;
	} | null>(null);

	const { data: peopleLabels } = useSuspenseQuery({
		queryKey: keys.peopleLabels.list(),
		queryFn: listPeopleLabels,
	});

	const groupColorMap = useMemo(() => {
		const m = new Map<string, string>();
		for (const l of peopleLabels ?? []) m.set(l.name, l.color);
		return m;
	}, [peopleLabels]);

	const graphData = useMemo(() => cloneGraphData(data), [data]);

	const groups = useMemo(
		() => [...new Set(data.nodes.map((n) => n.group).filter(Boolean))].sort(),
		[data.nodes],
	);

	const relTypes = useMemo(() => {
		const s = new Set<string>();
		for (const l of data.links) {
			if (l.type) s.add(l.type);
			if (l.reverse_type && l.reverse_type !== l.type) s.add(l.reverse_type);
		}
		return [...s].sort();
	}, [data.links]);

	const selfColor = useMemo(() => {
		const self = data.nodes.find((n) => n.is_self);
		return self?.group ? (groupColorMap.get(self.group) ?? PRIMARY) : PRIMARY;
	}, [data.nodes, groupColorMap]);

	const dimmedNodeIds = useMemo(() => {
		const dimmed = new Set<number>();
		if (activeGroup !== null) {
			for (const n of data.nodes) if (n.group !== activeGroup) dimmed.add(n.id);
		}
		if (activeRelType !== null) {
			const connected = new Set<number>();
			for (const l of data.links) {
				if (l.type === activeRelType || l.reverse_type === activeRelType) {
					connected.add(l.source as number);
					connected.add(l.target as number);
				}
			}
			for (const n of data.nodes) if (!connected.has(n.id)) dimmed.add(n.id);
		}
		return dimmed;
	}, [data.nodes, data.links, activeGroup, activeRelType]);

	// Measure container width
	useEffect(() => {
		const el = containerRef.current;
		if (!el) return;
		const ro = new ResizeObserver(([entry]) => {
			if (entry) setWidth(entry.contentRect.width);
		});
		ro.observe(el);
		return () => ro.disconnect();
	}, []);

	// Force remeasure when sidebar expands — canvas fixed-px width can prevent
	// the container from shrinking on its own until we explicitly re-read it.
	// biome-ignore lint/correctness/useExhaustiveDependencies: sidebarCollapsed is the signal, not a value used inside
	useEffect(() => {
		const el = containerRef.current;
		if (!el) return;
		const t = setTimeout(() => setWidth(el.getBoundingClientRect().width), 50);
		return () => clearTimeout(t);
	}, [sidebarCollapsed]);

	// Configure d3 forces once on mount
	useEffect(() => {
		const fg = fgRef.current;
		if (!fg) return;
		fg.d3Force("charge")?.strength?.(-300);
		fg.d3Force("link")?.distance?.(80);
		fg.d3ReheatSimulation();
	}, []);

	// biome-ignore lint/correctness/useExhaustiveDependencies: intentional — re-zoom when data swaps
	useEffect(() => {
		const t = setTimeout(() => fgRef.current?.zoomToFit(400, 40), 50);
		return () => clearTimeout(t);
	}, [graphData]);

	useEffect(() => {
		if (focusNodeId == null) return;
		const t = setTimeout(() => {
			const node = graphData.nodes.find((n) => n.id === focusNodeId) as
				| (GraphNode & { x?: number; y?: number })
				| undefined;
			if (node?.x != null && node?.y != null)
				fgRef.current?.centerAt(node.x, node.y, 600);
		}, 600);
		return () => clearTimeout(t);
	}, [focusNodeId, graphData]);

	useEffect(() => {
		if (repaintKey > 0) fgRef.current?.d3ReheatSimulation();
	}, [repaintKey]);

	const onImageLoad = useCallback(() => setRepaintKey((k) => k + 1), []);

	const nodeCanvasObject = useCallback(
		(node: object, ctx: CanvasRenderingContext2D, globalScale: number) => {
			drawNode(
				node as GraphNode & { x?: number; y?: number },
				ctx,
				globalScale,
				{
					colorBy,
					dimmedNodeIds,
					groupColorMap,
					showAvatar,
					onImageLoad,
				},
			);
		},
		[colorBy, dimmedNodeIds, groupColorMap, showAvatar, onImageLoad],
	);

	const nodePointerAreaPaint = useCallback(
		(
			node: object,
			color: string,
			ctx: CanvasRenderingContext2D,
			gs: number,
		) => {
			drawHitArea(
				node as GraphNode & { x?: number; y?: number },
				ctx,
				gs,
				color,
			);
		},
		[],
	);

	const getLinkColor = useCallback(
		(link: object) => {
			const l = link as GraphLink & {
				source?: GraphNode | number;
				target?: GraphNode | number;
			};
			const srcId =
				typeof l.source === "object" && l.source
					? (l.source as GraphNode).id
					: (l.source as number);
			const tgtId =
				typeof l.target === "object" && l.target
					? (l.target as GraphNode).id
					: (l.target as number);
			if (dimmedNodeIds.has(srcId) && dimmedNodeIds.has(tgtId))
				return "rgba(228,228,231,0.15)";
			if (
				activeRelType !== null &&
				l.type !== activeRelType &&
				(l.reverse_type ?? "") !== activeRelType
			)
				return "rgba(228,228,231,0.15)";
			return linkColor(l.type ?? "", colorBy);
		},
		[colorBy, dimmedNodeIds, activeRelType],
	);

	const buildConnections = useCallback(
		(nodeId: number) =>
			data.links
				.filter(
					(l) =>
						(l.source as number) === nodeId || (l.target as number) === nodeId,
				)
				.map((l) => {
					const isSource = (l.source as number) === nodeId;
					const otherId = isSource
						? (l.target as number)
						: (l.source as number);
					const other = data.nodes.find((n) => n.id === otherId);
					return {
						type: isSource ? l.type : l.reverse_type || l.type,
						otherName: other?.name ?? `#${otherId}`,
					};
				}),
		[data.links, data.nodes],
	);

	const handleNodeClick = useCallback(
		(nodeObj: object, event: MouseEvent) => {
			const n = nodeObj as GraphNode;
			const rect = containerRef.current?.getBoundingClientRect();
			if (!rect) return;
			setProfileCard({
				node: n,
				posX: event.clientX - rect.left,
				posY: event.clientY - rect.top,
			});
			const info: SelectedNodeInfo = {
				node: n,
				connections: buildConnections(n.id),
			};
			onNodeSelect?.(info);
		},
		[buildConnections, onNodeSelect],
	);

	const handleBackground = useCallback(() => {
		setProfileCard(null);
		onNodeSelect?.(null);
	}, [onNodeSelect]);

	const handleNodeHover = useCallback(
		(nodeObj: object | null, _prev: object | null, event?: MouseEvent) => {
			if (!nodeObj) {
				setTooltip(null);
				return;
			}
			const n = nodeObj as GraphNode;
			const rect = containerRef.current?.getBoundingClientRect();
			if (!event || !rect) return;
			const relTypes = new Set<string>();
			for (const l of data.links) {
				if ((l.source as number) === n.id && l.type) relTypes.add(l.type);
				else if ((l.target as number) === n.id && l.reverse_type)
					relTypes.add(l.reverse_type);
			}
			setTooltip({
				name: n.name,
				group: n.group,
				isSelf: n.is_self,
				relTypes: [...relTypes],
				x: event.clientX - rect.left,
				y: event.clientY - rect.top,
			});
		},
		[data.links],
	);

	const handleNodeDragEnd = useCallback((node: object) => {
		const n = node as GraphNode & {
			x?: number;
			y?: number;
			fx?: number | null;
			fy?: number | null;
		};
		n.fx = n.x;
		n.fy = n.y;
	}, []);

	return (
		<div
			ref={containerRef}
			className="relative min-w-0 rounded-md border border-zinc-200 bg-white"
		>
			<GraphControls
				title={title}
				colorBy={colorBy}
				onColorByChange={setColorBy}
				groups={groups}
				activeGroup={activeGroup}
				onGroupChange={setActiveGroup}
				relTypes={relTypes}
				activeRelType={activeRelType}
				onRelTypeChange={setActiveRelType}
				showAvatar={showAvatar}
				onShowAvatarChange={setShowAvatar}
				showOnlyMine={showOnlyMine}
				onShowOnlyMineChange={onShowOnlyMineChange}
				onRecenter={() => fgRef.current?.zoomToFit(400, 40)}
			/>

			{/* Stage */}
			<div
				className="relative overflow-hidden"
				style={{ height, background: DOTTED_BG }}
			>
				<ForceGraph2D
					ref={fgRef}
					graphData={graphData}
					width={width}
					height={height}
					nodeCanvasObject={nodeCanvasObject}
					nodePointerAreaPaint={nodePointerAreaPaint}
					nodeVal={(node) => ((node as GraphNode).is_self ? 45 : 16)}
					nodeRelSize={4}
					linkColor={getLinkColor}
					linkWidth={1.5}
					linkDirectionalArrowLength={0}
					onNodeClick={handleNodeClick}
					onNodeHover={handleNodeHover}
					onNodeDragEnd={handleNodeDragEnd}
					onBackgroundClick={handleBackground}
					cooldownTicks={120}
					enableNodeDrag
					enableZoomInteraction
					enablePanInteraction
				/>
			</div>

			<GraphLegend
				colorBy={colorBy}
				groups={groups}
				groupColorMap={groupColorMap}
				selfColor={selfColor}
				relTypes={relTypes}
			/>

			{/* Profile card overlay — positioned relative to containerRef */}
			{profileCard && (
				<GraphNodeProfileCard
					node={profileCard.node}
					posX={profileCard.posX}
					posY={profileCard.posY}
					showAvatar={showAvatar}
					groupColorMap={groupColorMap}
					onClose={() => {
						setProfileCard(null);
						onNodeSelect?.(null);
					}}
				/>
			)}

			{/* Hover tooltip */}
			{tooltip && (
				<div
					className="pointer-events-none absolute z-10 max-w-[180px] rounded bg-zinc-900/85 px-2 py-1 text-[11px] text-white"
					style={{ left: tooltip.x + 8, top: tooltip.y - 36 }}
				>
					<span className="font-semibold">{tooltip.name}</span>
					<span className="text-zinc-400">
						{tooltip.isSelf
							? " · you"
							: tooltip.group
								? ` · ${tooltip.group}`
								: ""}
					</span>
					{tooltip.relTypes.length > 0 && (
						<div className="mt-0.5 text-zinc-400">
							{tooltip.relTypes.join(", ")}
						</div>
					)}
				</div>
			)}
		</div>
	);
}
