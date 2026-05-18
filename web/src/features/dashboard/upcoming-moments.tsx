import { CalendarDays } from "lucide-react";
import { useState } from "react";
import { Button } from "#/components/ui/button";
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
			className="xl:col-span-5"
		>
			{isLoading ? (
				<div className="space-y-2">
					{["moment-1", "moment-2", "moment-3", "moment-4"].map((key) => (
						<div
							key={key}
							className="h-16 rounded-base bg-slate-100 animate-pulse"
						/>
					))}
				</div>
			) : visibleMoments.length ? (
				<div className="space-y-2">
					{visibleMoments.map((moment) => (
						<div
							key={moment.id}
							className="rounded-base border-2 border-slate-100 bg-slate-50/70 p-3 transition-colors hover:border-teal-200 hover:bg-teal-50/60"
						>
							<div className="flex items-start justify-between gap-3">
								<div className="min-w-0">
									<p className="truncate text-sm font-heading text-slate-900">
										{moment.personName}
									</p>
									<p className="mt-1 text-xs font-base text-slate-500">
										{moment.label} · {moment.detail}
									</p>
								</div>
								<span className="shrink-0 text-xs font-heading text-teal-700">
									{moment.date}
								</span>
							</div>
						</div>
					))}
					{moments.length > 5 ? (
						<Button
							type="button"
							variant="neutral"
							className="w-full border-slate-200 bg-white text-slate-700 hover:bg-teal-50"
							onClick={() => setExpanded((value) => !value)}
						>
							{expanded ? "Show less" : `Show ${moments.length - 5} more`}
						</Button>
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
