import { CalendarDays } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
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
	const { t } = useTranslation("dashboard");
	const [expanded, setExpanded] = useState(false);
	const visibleMoments = expanded ? moments : moments.slice(0, 5);

	return (
		<DashboardCard
			title={t("moments.title")}
			subtitle={t("moments.subtitle")}
			icon={CalendarDays}
			onRefresh={onRefresh}
			isRefreshing={isRefreshing}
		>
			{isLoading ? (
				<div className="space-y-px">
					{["m1", "m2", "m3", "m4"].map((key) => (
						<div key={key} className="h-12 bg-chip animate-pulse rounded" />
					))}
				</div>
			) : visibleMoments.length ? (
				<div>
					{visibleMoments.map((moment) => (
						<div
							key={moment.id}
							className="py-3 border-b border-line-soft last:border-b-0 hover:bg-chip -mx-4 px-4 transition-colors"
						>
							<div className="flex items-center justify-between gap-3">
								<div className="min-w-0">
									<p className="truncate text-[13px] text-ink">
										{moment.personName}
									</p>
									<p className="text-[11px] text-sub mt-0.5">
										{moment.label} · {moment.detail}
									</p>
								</div>
								<span className="shrink-0 font-mono text-[10px] text-sub">
									{moment.date}
								</span>
							</div>
						</div>
					))}
					{moments.length > 5 ? (
						<button
							type="button"
							className="w-full py-2 text-[11px] text-sub hover:bg-chip border-t border-line -mx-4 px-4 mt-1 transition-colors"
							onClick={() => setExpanded((v) => !v)}
						>
							{expanded
								? t("showLess")
								: t("showMore", { count: moments.length - 5 })}
						</button>
					) : null}
				</div>
			) : (
				<EmptyState
					icon={CalendarDays}
					title={t("moments.emptyTitle")}
					description={t("moments.emptyDescription")}
				/>
			)}
		</DashboardCard>
	);
}
