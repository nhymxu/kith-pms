import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { lazy, Suspense, useMemo, useState } from "react";
import type { RelationshipGraph } from "#/endpoints/relationships";
import { getRelationshipGraph } from "#/endpoints/relationships";
import type { SelectedNodeInfo } from "#/features/network/graph-selected-panel";
import { GraphSelectedPanel } from "#/features/network/graph-selected-panel";
import { getNetworkPrefs } from "#/lib/format-datetime";
import { keys } from "#/query-keys";

const LazyRelationshipGraph = lazy(
	() => import("#/features/network/relationship-graph"),
);

export const Route = createFileRoute("/_authed/network/")({
	component: NetworkPage,
	errorComponent: () => (
		<p className="py-4 text-[13px] text-red-600">
			Failed to load network graph.
		</p>
	),
});

interface NetworkGraphProps {
	showOnlyMine: boolean;
	onShowOnlyMineChange: (v: boolean) => void;
	onNodeSelect: (info: SelectedNodeInfo | null) => void;
	panelCollapsed: boolean;
}

function NetworkGraph({
	showOnlyMine,
	onShowOnlyMineChange,
	onNodeSelect,
	panelCollapsed,
}: NetworkGraphProps) {
	const { data } = useSuspenseQuery({
		queryKey: keys.relationships.graph(),
		queryFn: () => getRelationshipGraph(),
	});

	const selfNode = data.nodes.find((n) => n.is_self);

	const displayData = useMemo((): RelationshipGraph => {
		if (!showOnlyMine || !selfNode) return data;
		const included = new Set<number>([selfNode.id]);
		for (const link of data.links) {
			const src = link.source as number;
			const tgt = link.target as number;
			if (src === selfNode.id || tgt === selfNode.id) {
				included.add(src);
				included.add(tgt);
			}
		}
		return {
			nodes: data.nodes.filter((n) => included.has(n.id)),
			links: data.links.filter(
				(l) =>
					included.has(l.source as number) && included.has(l.target as number),
			),
		};
	}, [data, showOnlyMine, selfNode]);

	if (displayData.nodes.length === 0) {
		return (
			<div className="flex items-center justify-center rounded-md border border-zinc-200 bg-zinc-50 py-16 text-[13px] text-zinc-400">
				No connections yet. Add relationships to people to see your network.
			</div>
		);
	}

	return (
		<LazyRelationshipGraph
			data={displayData}
			// title="Your Network"
			height={Math.max(500, window.innerHeight - 200)}
			showOnlyMine={showOnlyMine}
			onShowOnlyMineChange={onShowOnlyMineChange}
			onNodeSelect={onNodeSelect}
			sidebarCollapsed={panelCollapsed}
		/>
	);
}

function NetworkPage() {
	const [showOnlyMine, setShowOnlyMine] = useState(
		() => getNetworkPrefs().networkShowOnlyMine,
	);
	const [selectedInfo, setSelectedInfo] = useState<SelectedNodeInfo | null>(
		null,
	);
	const [panelCollapsed, setPanelCollapsed] = useState(false);

	const fallback = (
		<div className="flex items-center justify-center rounded-md border border-zinc-200 bg-zinc-50 py-16 text-[13px] text-zinc-400">
			Loading network…
		</div>
	);

	return (
		<div className="space-y-4">
			<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">
				Your Network
			</h1>

			<div
				className={`grid grid-cols-1 gap-4 lg:items-start ${
					panelCollapsed
						? "lg:grid-cols-[1fr_auto]"
						: "lg:grid-cols-[1fr_280px]"
				}`}
			>
				<Suspense fallback={fallback}>
					<NetworkGraph
						showOnlyMine={showOnlyMine}
						onShowOnlyMineChange={setShowOnlyMine}
						onNodeSelect={setSelectedInfo}
						panelCollapsed={panelCollapsed}
					/>
				</Suspense>

				<GraphSelectedPanel
					selected={selectedInfo}
					collapsed={panelCollapsed}
					onCollapsedChange={setPanelCollapsed}
				/>
			</div>
		</div>
	);
}
