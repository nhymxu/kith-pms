import { formatBirthdayLabel, formatRelativeDate } from "./graph-date-format";
import type { GraphNode } from "./graph-types";

export interface SelectedNodeInfo {
	node: GraphNode;
	connections: Array<{ type: string; otherName: string }>;
}

interface GraphSelectedPanelProps {
	selected: SelectedNodeInfo | null;
	collapsed: boolean;
	onCollapsedChange: (v: boolean) => void;
}

function ChevronIcon({ direction }: { direction: "left" | "right" }) {
	return (
		<svg
			aria-hidden="true"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			strokeWidth={2}
			strokeLinecap="round"
			strokeLinejoin="round"
			className="h-3.5 w-3.5"
		>
			{direction === "left" ? (
				<path d="M15 18l-6-6 6-6" />
			) : (
				<path d="M9 18l6-6-6-6" />
			)}
		</svg>
	);
}

export function GraphSelectedPanel({
	selected,
	collapsed,
	onCollapsedChange,
}: GraphSelectedPanelProps) {
	if (collapsed) {
		return (
			<button
				type="button"
				onClick={() => onCollapsedChange(false)}
				title="Expand panel"
				aria-label="Expand selected panel"
				className="flex h-9 w-9 shrink-0 items-center justify-center self-start rounded-md border border-zinc-200 bg-white text-zinc-400 hover:border-indigo-400 hover:text-indigo-600"
			>
				<ChevronIcon direction="left" />
			</button>
		);
	}

	return (
		<div className="rounded-md border border-zinc-200 bg-white p-4">
			<div className="mb-3 flex items-center justify-between">
				<h2 className="text-[10px] font-semibold uppercase tracking-wider text-zinc-400">
					Selected
				</h2>
				<button
					type="button"
					onClick={() => onCollapsedChange(true)}
					title="Collapse panel"
					aria-label="Collapse selected panel"
					className="flex h-6 w-6 items-center justify-center rounded text-zinc-400 hover:bg-zinc-100 hover:text-zinc-600"
				>
					<ChevronIcon direction="right" />
				</button>
			</div>

			{!selected ? (
				<p className="text-[13px] text-zinc-400">
					Click any node to inspect it.
				</p>
			) : (
				<div className="space-y-3">
					{/* Name + group */}
					<div>
						<span className="text-[14px] font-semibold text-zinc-900">
							{selected.node.name}
						</span>
						{selected.node.group && (
							<span className="text-[13px] text-zinc-500">
								{" "}
								· {selected.node.group}
							</span>
						)}
					</div>

					{/* Birthday */}
					{selected.node.date_of_birth && (
						<div className="flex items-center gap-1.5 text-[12px] text-zinc-500">
							<span>🎂</span>
							<span>{formatBirthdayLabel(selected.node.date_of_birth)}</span>
						</div>
					)}

					{/* Last contacted */}
					{selected.node.last_contact_at && (
						<div className="flex items-center gap-1.5 text-[12px] text-zinc-500">
							<span>🕐</span>
							<span>
								Last contacted{" "}
								{formatRelativeDate(selected.node.last_contact_at)}
							</span>
						</div>
					)}

					{/* Connections list */}
					{selected.connections.length > 0 && (
						<div className="text-[12px] text-zinc-600">
							<span className="font-medium">
								{selected.connections.length} connection
								{selected.connections.length !== 1 ? "s" : ""}
							</span>
							<ul className="mt-1.5 space-y-1">
								{selected.connections.map((c, i) => (
									// biome-ignore lint/suspicious/noArrayIndexKey: stable ordered list
									<li key={i} className="flex items-baseline gap-1.5">
										<span className="shrink-0 rounded bg-zinc-100 px-1 py-0.5 text-[10px] font-medium text-zinc-500">
											{c.type}
										</span>
										<span className="text-zinc-700">{c.otherName}</span>
									</li>
								))}
							</ul>
						</div>
					)}
				</div>
			)}
		</div>
	);
}
