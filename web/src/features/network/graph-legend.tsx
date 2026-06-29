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
	return (
		<div className="flex flex-wrap items-center gap-x-4 gap-y-1 border-t border-zinc-200 px-4 py-2.5 text-[11px] text-zinc-500">
			{colorBy === "labels" ? (
				<>
					<span className="font-semibold text-zinc-700">Nodes by group:</span>
					<span className="flex items-center gap-1.5">
						<span
							className="h-2.5 w-2.5 rounded-full"
							style={{ background: selfColor }}
						/>
						You
					</span>
					{groups.map((g) => (
						<span key={g} className="flex items-center gap-1.5">
							<span
								className="h-2.5 w-2.5 rounded-full"
								style={{ background: groupColorMap.get(g) ?? "#71717a" }}
							/>
							{g}
						</span>
					))}
					<span>· edges: relationships</span>
				</>
			) : (
				<>
					<span className="font-semibold text-zinc-700">Edges by type:</span>
					{relTypes.map((t) => (
						<span key={t} className="flex items-center gap-1.5">
							<span
								className="h-0.5 w-4 rounded"
								style={{ background: typeColor(t) }}
							/>
							{t}
						</span>
					))}
				</>
			)}
		</div>
	);
}
