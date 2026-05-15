import { useQuery } from "@tanstack/react-query"
import { Users, BookOpen, Bell, Calendar } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "#/components/ui/card"
import { keys } from "#/query-keys"
import { listPeople } from "#/endpoints/people"
import { listJournal } from "#/endpoints/journal"
import { listReminders } from "#/endpoints/reminders"
import { listUpcomingDates } from "#/endpoints/dates"

function StatCard({
	icon: Icon,
	label,
	value,
	isLoading,
}: {
	icon: React.ElementType
	label: string
	value: number | undefined
	isLoading: boolean
}) {
	return (
		<Card>
			<CardHeader className="pb-2">
				<CardTitle className="text-sm font-base text-foreground/60 flex items-center gap-2">
					<Icon className="size-4" />
					{label}
				</CardTitle>
			</CardHeader>
			<CardContent>
				{isLoading ? (
					<div className="h-8 w-16 bg-border/30 rounded-base animate-pulse" />
				) : (
					<p className="text-3xl font-heading">{value ?? 0}</p>
				)}
			</CardContent>
		</Card>
	)
}

export function StatCards() {
	const people = useQuery({
		queryKey: keys.people.list(),
		queryFn: () => listPeople({ page_size: 1 }),
	})

	const journal = useQuery({
		queryKey: keys.journal.list({ page_size: 1 }),
		queryFn: () => listJournal({ page_size: 1 }),
	})

	const reminders = useQuery({
		queryKey: keys.reminders.list({ status: "upcoming" } as never),
		queryFn: () => listReminders({ status: "upcoming" }),
	})

	const dates = useQuery({
		queryKey: keys.dates.upcoming(),
		queryFn: () => listUpcomingDates(30),
	})

	return (
		<div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
			<StatCard
				icon={Users}
				label="People"
				value={people.data?.total}
				isLoading={people.isLoading}
			/>
			<StatCard
				icon={BookOpen}
				label="Journal entries"
				value={journal.data?.total}
				isLoading={journal.isLoading}
			/>
			<StatCard
				icon={Bell}
				label="Upcoming reminders"
				value={reminders.data?.length}
				isLoading={reminders.isLoading}
			/>
			<StatCard
				icon={Calendar}
				label="Upcoming dates"
				value={dates.data?.length}
				isLoading={dates.isLoading}
			/>
		</div>
	)
}
