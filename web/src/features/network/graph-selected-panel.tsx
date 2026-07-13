// Soft overlap: pending plan 260629-1555-relationship-network-visualization also
// touches features/network/*. Re-audit for reintroduced palette classes when
// that plan lands.

import { formatPersonName } from "#/lib/format-person-name";
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
				className="flex h-9 w-9 shrink-0 items-center justify-center self-start rounded-base border-bw border-line bg-panel text-sub hover:border-accent hover:text-accent-text"
			>
				<ChevronIcon direction="left" />
			</button>
		);
	}

	return (
		<div className="rounded-base border-bw border-line bg-panel p-4">
			<div className="mb-3 flex items-center justify-between">
				<h2 className="text-[10px] font-semibold uppercase tracking-wider text-sub">
					Selected
				</h2>
				<button
					type="button"
					onClick={() => onCollapsedChange(true)}
					title="Collapse panel"
					aria-label="Collapse selected panel"
					className="flex h-6 w-6 items-center justify-center rounded-base text-sub hover:bg-chip hover:text-ink"
				>
					<ChevronIcon direction="right" />
				</button>
			</div>

			{!selected ? (
				<p className="text-[13px] text-sub">Click any node to inspect it.</p>
			) : (
				<div className="space-y-3">
					{/* Name + nickname + group */}
					<div>
						<div className="text-[14px] font-semibold text-ink">
							{formatPersonName(selected.node.name, selected.node.nickname)}
						</div>
						{selected.node.groups && selected.node.groups.length > 0 && (
							<div className="flex flex-wrap gap-1">
								{selected.node.groups.map((g) => (
									<span
										key={g}
										className="rounded-base bg-chip px-1 py-0.5 text-[10px] font-medium text-chip-fg"
									>
										{g}
									</span>
								))}
							</div>
						)}
					</div>

					{/* Birthday */}
					{selected.node.date_of_birth && (
						<div className="flex items-center gap-1.5 text-[12px] text-sub">
							<span>🎂</span>
							<span>{formatBirthdayLabel(selected.node.date_of_birth)}</span>
						</div>
					)}

					{/* Last contacted */}
					{selected.node.last_contact_at && (
						<div className="flex items-center gap-1.5 text-[12px] text-sub">
							<span>🕐</span>
							<span>
								Last contacted{" "}
								{formatRelativeDate(selected.node.last_contact_at)}
							</span>
						</div>
					)}

					{/* Connections list */}
					{selected.connections.length > 0 && (
						<div className="text-[12px] text-sub">
							<span className="font-medium">
								{selected.connections.length} connection
								{selected.connections.length !== 1 ? "s" : ""}
							</span>
							<ul className="mt-1.5 space-y-1">
								{selected.connections.map((c, i) => (
									// biome-ignore lint/suspicious/noArrayIndexKey: stable ordered list
									<li key={i} className="flex items-baseline gap-1.5">
										<span className="shrink-0 rounded-base bg-chip px-1 py-0.5 text-[10px] font-medium text-chip-fg">
											{c.type}
										</span>
										<span className="text-ink">{c.otherName}</span>
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
