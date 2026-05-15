import { useQuery } from "@tanstack/react-query"
import { Link } from "@tanstack/react-router"
import { BookOpen } from "lucide-react"
import { Card, CardContent, CardHeader, CardTitle } from "#/components/ui/card"
import { Badge } from "#/components/ui/badge"
import { keys } from "#/query-keys"
import { listJournal } from "#/endpoints/journal"
import type { JournalActivity } from "#/schemas/journal"

function JournalCard({ entry }: { entry: JournalActivity }) {
	return (
		<Link
			to="/journal/$entryId"
			params={{ entryId: String(entry.id) }}
			className="block p-3 border-2 border-border rounded-base hover:translate-x-[2px] hover:translate-y-[2px] transition-transform"
		>
			<div className="flex items-start justify-between gap-2">
				<p className="text-sm font-heading truncate flex-1">{entry.title}</p>
				<span className="text-xs font-base text-foreground/50 shrink-0">{entry.occurred_at_date}</span>
			</div>
			{entry.people.length > 0 && (
				<div className="mt-1.5 flex flex-wrap gap-1">
					{entry.people.slice(0, 3).map((p) => (
						<Badge key={p.person_id} variant="neutral" className="text-xs">
							{p.name}
						</Badge>
					))}
					{entry.people.length > 3 && (
						<Badge variant="neutral" className="text-xs">+{entry.people.length - 3}</Badge>
					)}
				</div>
			)}
		</Link>
	)
}

export function RecentJournalActivity() {
	const { data, isLoading } = useQuery({
		queryKey: keys.journal.list({ page_size: 5 }),
		queryFn: () => listJournal({ page_size: 5 }),
	})

	return (
		<Card>
			<CardHeader className="pb-2">
				<CardTitle className="text-sm font-base text-foreground/60 flex items-center gap-2">
					<BookOpen className="size-4" />
					Recent journal entries
				</CardTitle>
			</CardHeader>
			<CardContent className="space-y-2">
				{isLoading &&
					Array.from({ length: 3 }).map((_, i) => (
						<div key={i} className="h-14 bg-border/20 rounded-base animate-pulse" />
					))}
				{!isLoading && (!data?.items.length) && (
					<p className="text-sm font-base text-foreground/50 py-4 text-center">No journal entries yet.</p>
				)}
				{data?.items.map((entry) => <JournalCard key={entry.id} entry={entry} />)}
			</CardContent>
		</Card>
	)
}
