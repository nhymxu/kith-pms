import { createFileRoute, Link, useNavigate } from "@tanstack/react-router"
import { useQuery } from "@tanstack/react-query"
import { z } from "zod"
import { listJournal } from "#/endpoints/journal"
import { keys } from "#/query-keys"
import { JournalTimeline } from "#/features/journal/journal-timeline"
import { Button } from "#/components/ui/button"

const searchSchema = z.object({
	q: z.string().optional(),
	page: z.coerce.number().min(1).optional().default(1),
	page_size: z.coerce.number().min(1).max(100).optional().default(20),
	from_date: z.string().optional(),
	to_date: z.string().optional(),
})

export const Route = createFileRoute("/_authed/journal/")({
	validateSearch: searchSchema,
	component: JournalPage,
})

function JournalPage() {
	const navigate = useNavigate()
	const search = Route.useSearch()

	const { data, isPending, isError } = useQuery({
		queryKey: keys.journal.list({ page: search.page, page_size: search.page_size }),
		queryFn: () => listJournal({ q: search.q, page: search.page, page_size: search.page_size, from_date: search.from_date, to_date: search.to_date }),
	})

	if (isError) return <p className="text-sm font-base text-destructive">Failed to load journal.</p>

	return (
		<div className="space-y-4">
			<div className="flex items-center justify-between">
				<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">Journal</h1>
				<Button asChild>
					<Link to="/journal/new">New Entry</Link>
				</Button>
			</div>

			{/* Date range filter */}
			<div className="flex flex-wrap gap-3 items-end">
				<div className="space-y-1">
					<label className="text-[11px] font-medium text-zinc-500">From</label>
					<input
						type="date"
						value={search.from_date ?? ""}
						onChange={(e) => void navigate({ to: "/journal", search: { ...search, from_date: e.target.value || undefined, page: 1 } })}
						className="h-9 border border-zinc-200 rounded-md bg-white px-3 text-[13px] focus:outline-none focus:ring-2 focus:ring-indigo-600"
					/>
				</div>
				<div className="space-y-1">
					<label className="text-[11px] font-medium text-zinc-500">To</label>
					<input
						type="date"
						value={search.to_date ?? ""}
						onChange={(e) => void navigate({ to: "/journal", search: { ...search, to_date: e.target.value || undefined, page: 1 } })}
						className="h-9 border border-zinc-200 rounded-md bg-white px-3 text-[13px] focus:outline-none focus:ring-2 focus:ring-indigo-600"
					/>
				</div>
				{(search.from_date || search.to_date) && (
					<Button
						variant="neutral"
						size="sm"
						onClick={() => void navigate({ to: "/journal", search: { ...search, from_date: undefined, to_date: undefined, page: 1 } })}
					>
						Clear dates
					</Button>
				)}
			</div>

			{isPending ? (
				<p className="text-sm font-base text-foreground/60 py-4">Loading…</p>
			) : (
				<JournalTimeline data={data?.items ?? []} />
			)}
		</div>
	)
}
