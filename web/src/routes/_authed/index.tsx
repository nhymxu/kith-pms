import { createFileRoute } from "@tanstack/react-router"
import { StatCards } from "#/features/dashboard/stat-cards"
import { RecentJournalActivity } from "#/features/dashboard/recent-journal-activity"

export const Route = createFileRoute("/_authed/")({
	component: DashboardPage,
})

function DashboardPage() {
	return (
		<div className="space-y-6">
			<h1 className="text-2xl font-heading">Dashboard</h1>
			<StatCards />
			<RecentJournalActivity />
		</div>
	)
}
