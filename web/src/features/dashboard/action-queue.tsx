import { Bell, Gift } from "lucide-react";
import { useMemo, useState } from "react";
import { Badge } from "#/components/ui/badge";
import { Button } from "#/components/ui/button";
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
			className="xl:col-span-5"
		>
			<div className="mb-4 flex gap-2 overflow-x-auto pb-1">
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
				<div className="space-y-2">
					{["action-1", "action-2", "action-3", "action-4", "action-5"].map(
						(key) => (
							<div
								key={key}
								className="h-16 rounded-base bg-slate-100 animate-pulse"
							/>
						),
					)}
				</div>
			) : visibleActions.length ? (
				<div className="space-y-2">
					{visibleActions.map((action) => (
						<div
							key={action.id}
							className="rounded-base border-2 border-slate-100 bg-slate-50/70 p-3 transition-colors hover:border-teal-200 hover:bg-teal-50/60"
						>
							<div className="flex items-start justify-between gap-3">
								<div className="min-w-0">
									<p className="truncate text-sm font-heading text-slate-900">
										{action.label}
									</p>
									<p className="mt-1 line-clamp-1 text-xs font-base text-slate-500">
										{action.personName ? `${action.personName} · ` : ""}
										{action.detail}
									</p>
								</div>
								<Badge variant="neutral" className={badgeClass(action.type)}>
									{itemLabel(action.type)}
								</Badge>
							</div>
							<p className="mt-2 text-xs font-base text-slate-500">
								{action.date}
							</p>
						</div>
					))}
					{filteredActions.length > 8 ? (
						<Button
							type="button"
							variant="neutral"
							className="w-full border-slate-200 bg-white text-slate-700 hover:bg-teal-50"
							onClick={() => setExpanded((value) => !value)}
						>
							{expanded
								? "Show less"
								: `Show ${filteredActions.length - 8} more`}
						</Button>
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

function badgeClass(type: DashboardAction["type"]): string {
	if (type === "overdue") return "border-red-200 bg-red-50 text-red-700";
	if (type === "today") return "border-teal-200 bg-teal-50 text-teal-700";
	if (type === "gift") return "border-amber-200 bg-amber-50 text-amber-700";
	return "border-slate-200 bg-white text-slate-600";
}
