// Upcoming dates list: grouped by month, shows person name, date type, days-until badge
import { useQuery } from "@tanstack/react-query"
import { Link } from "@tanstack/react-router"
import { keys } from "#/query-keys"
import { listUpcomingDates } from "#/endpoints/dates"
import { Badge } from "#/components/ui/badge"
import { Card, CardContent } from "#/components/ui/card"

function daysUntil(dateStr: string): number {
	const today = new Date()
	today.setHours(0, 0, 0, 0)
	const target = new Date(dateStr)
	target.setHours(0, 0, 0, 0)
	return Math.round((target.getTime() - today.getTime()) / (1000 * 60 * 60 * 24))
}

function monthLabel(dateStr: string): string {
	const d = new Date(dateStr)
	return d.toLocaleDateString(undefined, { month: "long", year: "numeric" })
}

function DaysUntilBadge({ days }: { days: number }) {
	if (days === 0) return <Badge className="bg-main text-main-foreground">Today</Badge>
	if (days <= 7) return <Badge className="bg-yellow-300 text-black border-black">In {days}d</Badge>
	if (days <= 30) return <Badge variant="neutral">In {days}d</Badge>
	return <Badge variant="neutral" className="text-foreground/50">In {days}d</Badge>
}

export function DatesList() {
	const { data, isPending, isError } = useQuery({
		queryKey: keys.dates.upcoming(),
		queryFn: () => listUpcomingDates(90),
	})

	if (isPending) return <p className="text-sm font-base text-foreground/60">Loading upcoming dates…</p>
	if (isError) return <p className="text-sm font-base text-destructive">Failed to load dates.</p>
	if (!data?.length) return <p className="text-sm font-base text-foreground/50">No upcoming dates in the next 90 days.</p>

	// Group by month of next_occurrence
	const groups = new Map<string, typeof data>()
	for (const item of data) {
		const key = monthLabel(item.next_occurrence)
		if (!groups.has(key)) groups.set(key, [])
		groups.get(key)!.push(item)
	}

	return (
		<div className="space-y-6">
			{Array.from(groups.entries()).map(([month, items]) => (
				<div key={month}>
					<h2 className="text-sm font-heading uppercase tracking-wide text-foreground/60 mb-2">{month}</h2>
					<Card>
						<CardContent className="p-0 divide-y-2 divide-border">
							{items.map((item, i) => {
								const days = daysUntil(item.next_occurrence)
								return (
									<div key={i} className="flex items-center gap-3 px-4 py-3">
										<div className="flex-1 min-w-0">
											<Link
												to="/people/$personId"
												params={{ personId: String(item.person.id) }}
												className="font-heading text-sm hover:underline"
											>
												{item.person.name}
											</Link>
											<p className="text-xs text-foreground/60 font-base capitalize">
												{item.kind}{item.years_since > 0 ? ` · ${item.years_since + 1} years` : ""}
											</p>
										</div>
										<DaysUntilBadge days={days} />
									</div>
								)
							})}
						</CardContent>
					</Card>
				</div>
			))}
		</div>
	)
}
