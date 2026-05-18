import { CalendarDays } from "lucide-react";
import { useState } from "react";
import { DashboardCard } from "./dashboard-card";
import type { DashboardMoment } from "./dashboard-data";
import { EmptyState } from "./empty-state";

export function UpcomingMoments({
	moments,
	isLoading,
	onRefresh,
	isRefreshing,
}: {
	moments: DashboardMoment[];
	isLoading: boolean;
	onRefresh: () => void;
	isRefreshing: boolean;
}) {
	const [expanded, setExpanded] = useState(false);
	const visibleMoments = expanded ? moments : moments.slice(0, 5);

	return (
		<DashboardCard
			title="Upcoming moments"
			subtitle="Dates worth preparing for"
			icon={CalendarDays}
			onRefresh={onRefresh}
			isRefreshing={isRefreshing}
		>
			{isLoading ? (
				<div className="space-y-px">
					{["m1", "m2", "m3", "m4"].map((key) => (
						<div key={key} className="h-12 bg-zinc-100 animate-pulse rounded" />
					))}
				</div>
			) : visibleMoments.length ? (
				<div>
					{visibleMoments.map((moment) => (
						<div
							key={moment.id}
							className="py-3 border-b border-zinc-100 last:border-b-0 hover:bg-zinc-50 -mx-4 px-4 transition-colors"
						>
							<div className="flex items-center justify-between gap-3">
								<div className="min-w-0">
									<p className="truncate text-[13px] text-zinc-900">{moment.personName}</p>
									<p className="text-[11px] text-zinc-500 mt-0.5">
										{moment.label} · {moment.detail}
									</p>
								</div>
								<span className="shrink-0 font-mono text-[10px] text-zinc-500">{moment.date}</span>
							</div>
						</div>
					))}
					{moments.length > 5 ? (
						<button
							type="button"
							className="w-full py-2 text-[11px] text-zinc-600 hover:bg-zinc-50 border-t border-zinc-200 -mx-4 px-4 mt-1 transition-colors"
							onClick={() => setExpanded((v) => !v)}
						>
							{expanded ? "Show less" : `Show ${moments.length - 5} more`}
						</button>
					) : null}
				</div>
			) : (
				<EmptyState
					icon={CalendarDays}
					title="No upcoming moments"
					description="Birthdays and milestones in the next 30 days appear here."
				/>
			)}
		</DashboardCard>
	);
}
