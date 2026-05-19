// Upcoming dates list: grouped by month, shows person name, date type, days-until badge
import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { listUpcomingDates } from "#/endpoints/dates";
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

function DaysUntilBadge({ days }: { days: number }) {
	if (days === 0)
		return (
			<span className="font-mono text-[11px] text-indigo-600 font-medium">
				Today
			</span>
		);
	if (days <= 7)
		return (
			<span className="font-mono text-[11px] text-amber-600">In {days}d</span>
		);
	if (days <= 30)
		return (
			<span className="font-mono text-[11px] text-zinc-500">In {days}d</span>
		);
	return (
		<span className="font-mono text-[11px] text-zinc-400">In {days}d</span>
	);
}

export function DatesList() {
	const { data, isPending, isError } = useQuery({
		queryKey: keys.dates.upcoming(),
		queryFn: () => listUpcomingDates(90),
	});

	if (isPending)
		return <p className="text-[13px] text-zinc-500">Loading upcoming dates…</p>;
	if (isError)
		return <p className="text-[13px] text-red-600">Failed to load dates.</p>;
	if (!data?.length)
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
		groups.get(key)!.push(item);
	}

	return (
		<div className="space-y-6">
			{Array.from(groups.entries()).map(([month, items]) => (
				<div key={month}>
					<h2 className="text-[11px] font-medium uppercase tracking-wider text-zinc-500 mb-2">
						{month}
					</h2>
					<div className="border border-zinc-200 rounded-md bg-white divide-y divide-zinc-100">
						{items.map((item, i) => {
							const days = daysUntil(item.next_occurrence);
							return (
								<div
									key={i}
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
									<DaysUntilBadge days={days} />
								</div>
							);
						})}
					</div>
				</div>
			))}
		</div>
	);
}
