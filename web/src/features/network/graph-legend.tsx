// Soft overlap: pending plan 260629-1555-relationship-network-visualization also
// touches features/network/*. Re-audit for reintroduced palette classes when
// that plan lands.

import { useTranslation } from "react-i18next";
import { readToken } from "#/lib/read-css-token";
import { typeColor } from "./graph-node-canvas";
import type { ColorBy } from "./graph-types";

interface GraphLegendProps {
	colorBy: ColorBy;
	groups: string[];
	groupColorMap: Map<string, string>;
	selfColor: string;
	relTypes: string[];
}

export function GraphLegend({
	colorBy,
	groups,
	groupColorMap,
	selfColor,
	relTypes,
}: GraphLegendProps) {
	const { t } = useTranslation("network");
	return (
		<div className="flex flex-wrap items-center gap-x-4 gap-y-1 border-t border-bw border-line px-4 py-2.5 text-[11px] text-sub">
			{colorBy === "labels" ? (
				<>
					<span className="font-semibold text-ink">
						{t("legend.nodesByGroup")}
					</span>
					<span className="flex items-center gap-1.5">
						<span
							className="h-2.5 w-2.5 rounded-full"
							style={{ background: selfColor }}
						/>
						{t("legend.you")}
					</span>
					{groups.map((g) => (
						<span key={g} className="flex items-center gap-1.5">
							<span
								className="h-2.5 w-2.5 rounded-full"
								style={{
									background:
										groupColorMap.get(g) ?? readToken("--sub", "#71717a"),
								}}
							/>
							{g}
						</span>
					))}
					<span>{t("legend.edgesAreRelationships")}</span>
				</>
			) : (
				<>
					<span className="font-semibold text-ink">
						{t("legend.edgesByType")}
					</span>
					{relTypes.map((rel) => (
						<span key={rel} className="flex items-center gap-1.5">
							<span
								className="h-0.5 w-4 rounded"
								style={{ background: typeColor(rel) }}
							/>
							{rel}
						</span>
					))}
				</>
			)}
		</div>
	);
}
