import { Bell, BookOpen, CalendarDays, Gift, Users } from "lucide-react";
import type { DashboardSummaryCard } from "./dashboard-data";

const icons = {
	people: Users,
	followups: Bell,
	dates: CalendarDays,
	gifts: Gift,
	journal: BookOpen,
} satisfies Record<DashboardSummaryCard["id"], typeof Users>;

const accentClass: Record<DashboardSummaryCard["id"], string> = {
	people: "text-zinc-500",
	followups: "text-red-600",
	dates: "text-indigo-600",
	gifts: "text-zinc-500",
	journal: "text-zinc-500",
};

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
		<div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 divide-x divide-zinc-200 border border-zinc-200 rounded-md bg-white">
			{cards.map((card) => {
				const Icon = icons[card.id];
				return (
					<div key={card.id} className="p-4 min-w-0">
						<p className="text-[10px] uppercase tracking-wider text-zinc-500 flex items-center gap-1.5">
							<Icon className="size-3 shrink-0" />
							{card.label}
						</p>
						{isLoading ? (
							<>
								<div className="h-6 w-12 rounded bg-zinc-100 animate-pulse mt-1" />
								<div className="h-3 w-20 rounded bg-zinc-100 animate-pulse mt-1.5" />
							</>
						) : (
							<>
								<p className="font-mono text-xl font-semibold text-zinc-900 mt-1">
									{card.value}
								</p>
								<p className={`text-[11px] mt-0.5 ${accentClass[card.id]} ${isStale ? "text-zinc-400" : ""}`}>
									{isStale ? "Cached" : card.detail}
								</p>
							</>
						)}
					</div>
				);
			})}
		</div>
	);
}
