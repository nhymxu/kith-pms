import { Bell, BookOpen, CalendarDays, Gift, Users } from "lucide-react";
import { DashboardCard } from "./dashboard-card";
import type { DashboardSummaryCard } from "./dashboard-data";

const icons = {
	people: Users,
	followups: Bell,
	dates: CalendarDays,
	gifts: Gift,
	journal: BookOpen,
} satisfies Record<DashboardSummaryCard["id"], typeof Users>;

export function SummaryCards({
	cards,
	isLoading,
	isStale,
	onRefresh,
	refreshingId,
}: {
	cards: DashboardSummaryCard[];
	isLoading: boolean;
	isStale: boolean;
	onRefresh: (id: DashboardSummaryCard["id"]) => void;
	refreshingId?: DashboardSummaryCard["id"];
}) {
	return (
		<div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-5">
			{cards.map((card) => (
				<DashboardCard
					key={card.id}
					title={card.label}
					subtitle={isStale ? "Showing cached data" : card.trend}
					icon={icons[card.id]}
					onRefresh={() => onRefresh(card.id)}
					isRefreshing={refreshingId === card.id}
				>
					{isLoading ? (
						<div className="h-16 rounded-base bg-slate-100 animate-pulse" />
					) : (
						<div className="space-y-2">
							<p className="text-3xl font-heading tracking-tight text-slate-950">
								{card.value}
							</p>
							<p className="text-sm font-base text-slate-500">{card.detail}</p>
						</div>
					)}
				</DashboardCard>
			))}
		</div>
	);
}
