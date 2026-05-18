import { Link } from "@tanstack/react-router";
import { BookOpen } from "lucide-react";
import { useState } from "react";
import { DashboardCard } from "./dashboard-card";
import type { DashboardActivity } from "./dashboard-data";
import { EmptyState } from "./empty-state";

export function RecentRelationshipActivity({
	activities,
	isLoading,
	onRefresh,
	isRefreshing,
}: {
	activities: DashboardActivity[];
	isLoading: boolean;
	onRefresh: () => void;
	isRefreshing: boolean;
}) {
	const [expanded, setExpanded] = useState(false);
	const visibleActivities = expanded ? activities : activities.slice(0, 6);

	return (
		<DashboardCard
			title="Recent activity"
			subtitle="Latest journal context"
			icon={BookOpen}
			onRefresh={onRefresh}
			isRefreshing={isRefreshing}
		>
			{isLoading ? (
				<div className="space-y-px">
					{["r1", "r2", "r3", "r4"].map((key) => (
						<div key={key} className="h-14 bg-zinc-100 animate-pulse rounded" />
					))}
				</div>
			) : visibleActivities.length ? (
				<div>
					{visibleActivities.map((activity) => (
						<Link
							key={activity.id}
							to="/journal/$entryId"
							params={{ entryId: activity.id }}
							className="block py-3 border-b border-zinc-100 last:border-b-0 hover:bg-zinc-50 -mx-4 px-4 transition-colors"
						>
							<div className="flex items-start justify-between gap-3">
								<div className="min-w-0">
									<p className="truncate text-[13px] text-zinc-900">{activity.title}</p>
									<p className="text-[11px] text-zinc-500 mt-0.5 line-clamp-1">{activity.detail}</p>
								</div>
								<span className="shrink-0 font-mono text-[10px] text-zinc-500">{activity.date}</span>
							</div>
							{activity.people.length ? (
								<div className="mt-1.5 flex flex-wrap gap-1">
									{activity.people.slice(0, 3).map((person) => (
										<span key={person} className="text-[10px] text-indigo-600">
											@{person}
										</span>
									))}
									{activity.people.length > 3 ? (
										<span className="text-[10px] text-zinc-400">
											+{activity.people.length - 3}
										</span>
									) : null}
								</div>
							) : null}
						</Link>
					))}
					{activities.length > 6 ? (
						<button
							type="button"
							className="w-full py-2 text-[11px] text-zinc-600 hover:bg-zinc-50 border-t border-zinc-200 -mx-4 px-4 mt-1 transition-colors"
							onClick={() => setExpanded((v) => !v)}
						>
							{expanded ? "Show less" : `Show ${activities.length - 6} more`}
						</button>
					) : null}
				</div>
			) : (
				<EmptyState
					icon={BookOpen}
					title="No journal activity"
					description="Log a recent interaction to build relationship context."
				/>
			)}
		</DashboardCard>
	);
}
