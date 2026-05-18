import { Link } from "@tanstack/react-router";
import { BookOpen } from "lucide-react";
import { useState } from "react";
import { Badge } from "#/components/ui/badge";
import { Button } from "#/components/ui/button";
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
			className="xl:col-span-7"
		>
			{isLoading ? (
				<div className="space-y-2">
					{["activity-1", "activity-2", "activity-3", "activity-4"].map(
						(key) => (
							<div
								key={key}
								className="h-16 rounded-base bg-slate-100 animate-pulse"
							/>
						),
					)}
				</div>
			) : visibleActivities.length ? (
				<div className="space-y-2">
					{visibleActivities.map((activity) => (
						<Link
							key={activity.id}
							to="/journal/$entryId"
							params={{ entryId: activity.id }}
							className="block rounded-base border-2 border-slate-100 bg-slate-50/70 p-3 transition-colors hover:border-teal-200 hover:bg-teal-50/60"
						>
							<div className="flex items-start justify-between gap-3">
								<div className="min-w-0">
									<p className="truncate text-sm font-heading text-slate-900">
										{activity.title}
									</p>
									<p className="mt-1 line-clamp-1 text-xs font-base text-slate-500">
										{activity.detail}
									</p>
								</div>
								<span className="shrink-0 text-xs font-base text-slate-500">
									{activity.date}
								</span>
							</div>
							{activity.people.length ? (
								<div className="mt-2 flex flex-wrap gap-1">
									{activity.people.slice(0, 3).map((person) => (
										<Badge
											key={person}
											variant="neutral"
											className="border-slate-200 bg-white text-xs text-slate-600"
										>
											{person}
										</Badge>
									))}
									{activity.people.length > 3 ? (
										<Badge
											variant="neutral"
											className="border-slate-200 bg-white text-xs text-slate-600"
										>
											+{activity.people.length - 3}
										</Badge>
									) : null}
								</div>
							) : null}
						</Link>
					))}
					{activities.length > 6 ? (
						<Button
							type="button"
							variant="neutral"
							className="w-full border-slate-200 bg-white text-slate-700 hover:bg-teal-50"
							onClick={() => setExpanded((value) => !value)}
						>
							{expanded ? "Show less" : `Show ${activities.length - 6} more`}
						</Button>
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
