import { createFileRoute, Link, useNavigate } from "@tanstack/react-router"
import { z } from "zod"
import { Button } from "#/components/ui/button"
import { PeopleTable } from "#/features/people/people-table"

const searchSchema = z.object({
	q: z.string().optional(),
	page: z.coerce.number().min(1).optional().default(1),
	page_size: z.coerce.number().min(1).max(100).optional().default(20),
	labels: z.array(z.coerce.number()).optional(),
})

export const Route = createFileRoute("/_authed/people/")({
	validateSearch: searchSchema,
	component: PeoplePage,
})

function PeoplePage() {
	const navigate = useNavigate()
	const search = Route.useSearch()

	function handleSearchChange(q: string) {
		void navigate({ to: "/people", search: { ...search, q: q || undefined, page: 1 } })
	}

	function handleLabelsChange(labels: number[]) {
		void navigate({ to: "/people", search: { ...search, labels: labels.length ? labels : undefined, page: 1 } })
	}

	return (
		<div className="space-y-4">
			<div className="flex items-center justify-between">
				<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">People</h1>
				<Button asChild>
					<Link to="/people/new">New Person</Link>
				</Button>
			</div>
			<PeopleTable
				q={search.q}
				labels={search.labels}
				page={search.page}
				page_size={search.page_size}
				onSearchChange={handleSearchChange}
				onLabelsChange={handleLabelsChange}
			/>
		</div>
	)
}
