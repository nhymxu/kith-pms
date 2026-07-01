// Upcoming dates list: grouped by month, shows person name, date type, exact date + days-until badge
import { useSuspenseQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { QueryBoundary } from "#/components/query-boundary";
import { listUpcomingDates } from "#/endpoints/important-dates";
import { keys } from "#/query-keys";

function daysUntil(dateStr: string): number {
	const today = new Date();
	today.setHours(0, 0, 0, 0);
	const target = new Date(dateStr);
	target.setHours(0, 0, 0, 0);
	return Math.round(
		(target.getTime() - today.getTime()) / (1000 * 60 * 60 * 24),
	);
}

function monthLabel(dateStr: string): string {
	const d = new Date(dateStr);
	return d.toLocaleDateString(undefined, { month: "long", year: "numeric" });
}

function exactDateLabel(dateStr: string): string {
	const d = new Date(dateStr);
	const currentYear = new Date().getFullYear();
	if (d.getFullYear() === currentYear) {
		return d.toLocaleDateString(undefined, { month: "short", day: "numeric" });
	}
	return d.toLocaleDateString(undefined, {
		month: "short",
		day: "numeric",
		year: "numeric",
	});
}

function DaysUntilBadge({ days, dateStr }: { days: number; dateStr: string }) {
	const exact = exactDateLabel(dateStr);
	if (days === 0)
		return (
			<div className="flex flex-col items-end gap-0.5">
				<span className="font-mono text-[11px] text-indigo-600 font-medium">
					Today
				</span>
				<span className="text-[10px] text-zinc-400">{exact}</span>
			</div>
		);
	if (days <= 7)
		return (
			<div className="flex flex-col items-end gap-0.5">
				<span className="font-mono text-[11px] text-amber-600">In {days}d</span>
				<span className="text-[10px] text-zinc-400">{exact}</span>
			</div>
		);
	if (days <= 30)
		return (
			<div className="flex flex-col items-end gap-0.5">
				<span className="font-mono text-[11px] text-zinc-500">In {days}d</span>
				<span className="text-[10px] text-zinc-400">{exact}</span>
			</div>
		);
	return (
		<div className="flex flex-col items-end gap-0.5">
			<span className="font-mono text-[11px] text-zinc-400">In {days}d</span>
			<span className="text-[10px] text-zinc-400">{exact}</span>
		</div>
	);
}

function DatesListInner() {
	const { data } = useSuspenseQuery({
		queryKey: keys.dates.upcoming(),
		queryFn: () => listUpcomingDates(90),
	});

	if (!data.length)
		return (
			<p className="text-[13px] text-zinc-500">
				No upcoming dates in the next 90 days.
			</p>
		);

	// Group by month of next_occurrence
	const groups = new Map<string, typeof data>();
	for (const item of data) {
		const key = monthLabel(item.next_occurrence);
		if (!groups.has(key)) groups.set(key, []);
		groups.get(key)?.push(item);
	}

	return (
		<div className="space-y-6">
			{Array.from(groups.entries()).map(([month, items]) => (
				<div key={month}>
					<h2 className="text-[11px] font-medium uppercase tracking-wider text-zinc-500 mb-2">
						{month}
					</h2>
					<div className="border border-zinc-200 rounded-md bg-white divide-y divide-zinc-100">
						{items.map((item, _i) => {
							const days = daysUntil(item.next_occurrence);
							return (
								<div
									key={`${item.person.id}-${item.kind}`}
									className="flex items-center gap-3 px-4 py-3 hover:bg-zinc-50 transition-colors"
								>
									<div className="flex-1 min-w-0">
										<Link
											to="/people/$personId"
											params={{ personId: String(item.person.id) }}
											className="text-[13px] text-zinc-900 hover:text-indigo-600 hover:underline"
										>
											{item.person.name}
										</Link>
										<p className="text-[11px] text-zinc-500 capitalize">
											{item.kind}
											{item.years_since > 0
												? ` · ${item.years_since + 1} years`
												: ""}
										</p>
									</div>
									<DaysUntilBadge days={days} dateStr={item.next_occurrence} />
								</div>
							);
						})}
					</div>
				</div>
			))}
		</div>
	);
}

const datesListFallback = (
	<p className="text-[13px] text-zinc-500">Loading upcoming dates…</p>
);

export function DatesList() {
	return (
		<QueryBoundary fallback={datesListFallback}>
			<DatesListInner />
		</QueryBoundary>
	);
}
