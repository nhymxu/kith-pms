import { Bell, Gift } from "lucide-react";
import { useMemo, useState } from "react";
import { DashboardCard } from "./dashboard-card";
import type { DashboardAction } from "./dashboard-data";
import { DashboardFilterPill } from "./dashboard-filter-pill";
import { EmptyState } from "./empty-state";

const filters = ["all", "overdue", "today", "upcoming", "gift"] as const;
type ActionFilter = (typeof filters)[number];

export function ActionQueue({
	actions,
	isLoading,
	onRefresh,
	isRefreshing,
}: {
	actions: DashboardAction[];
	isLoading: boolean;
	onRefresh: () => void;
	isRefreshing: boolean;
}) {
	const [filter, setFilter] = useState<ActionFilter>("all");
	const [expanded, setExpanded] = useState(false);
	const filteredActions = useMemo(
		() =>
			filter === "all"
				? actions
				: actions.filter((action) => action.type === filter),
		[actions, filter],
	);
	const visibleActions = expanded
		? filteredActions
		: filteredActions.slice(0, 8);

	return (
		<DashboardCard
			title="Action queue"
			subtitle="Follow-ups, reminders, and gift ideas"
			icon={Bell}
			onRefresh={onRefresh}
			isRefreshing={isRefreshing}
		>
			<div className="mb-3 flex gap-1.5 overflow-x-auto pb-1">
				{filters.map((item) => (
					<DashboardFilterPill
						key={item}
						label={itemLabel(item)}
						active={filter === item}
						onClick={() => setFilter(item)}
					/>
				))}
			</div>
			{isLoading ? (
				<div className="space-y-px">
					{["a1", "a2", "a3", "a4", "a5"].map((key) => (
						<div key={key} className="h-14 bg-zinc-100 animate-pulse rounded" />
					))}
				</div>
			) : visibleActions.length ? (
				<div>
					{visibleActions.map((action) => (
						<div
							key={action.id}
							className="px-0 py-3 border-b border-zinc-100 last:border-b-0 hover:bg-zinc-50 -mx-4 px-4 transition-colors"
						>
							<div className="flex items-start justify-between gap-3">
								<div className="min-w-0">
									<p className="truncate text-[13px] text-zinc-900">
										{action.label}
									</p>
									<p className="text-[11px] text-zinc-500 mt-0.5">
										{action.personName ? `${action.personName} · ` : ""}
										{action.detail}
									</p>
								</div>
								<span
									className={`font-mono text-[10px] uppercase shrink-0 ${statusClass(action.type)}`}
								>
									{itemLabel(action.type)}
								</span>
							</div>
						</div>
					))}
					{filteredActions.length > 8 ? (
						<button
							type="button"
							className="w-full py-2 text-[11px] text-zinc-600 hover:bg-zinc-50 border-t border-zinc-200 -mx-4 px-4 mt-1 transition-colors"
							onClick={() => setExpanded((v) => !v)}
						>
							{expanded
								? "Show less"
								: `Show ${filteredActions.length - 8} more`}
						</button>
					) : null}
				</div>
			) : (
				<EmptyState
					icon={filter === "gift" ? Gift : Bell}
					title="Queue is clear"
					description="No matching follow-ups need attention."
				/>
			)}
		</DashboardCard>
	);
}

function itemLabel(value: ActionFilter | DashboardAction["type"]): string {
	return value === "all" ? "All" : value[0].toUpperCase() + value.slice(1);
}

function statusClass(type: DashboardAction["type"]): string {
	if (type === "overdue") return "text-red-600";
	if (type === "today") return "text-indigo-600";
	if (type === "gift") return "text-amber-600";
	return "text-zinc-500";
}
