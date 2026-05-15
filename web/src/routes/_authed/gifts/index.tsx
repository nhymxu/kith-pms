import { createFileRoute, Link } from "@tanstack/react-router"
import { useQuery } from "@tanstack/react-query"
import { listGifts } from "#/endpoints/gifts"
import { keys } from "#/query-keys"
import { GiftsTable } from "#/features/gifts/gifts-table"
import { Button } from "#/components/ui/button"

export const Route = createFileRoute("/_authed/gifts/")({
	component: GiftsPage,
})

function GiftsPage() {
	const { data, isPending, isError } = useQuery({
		queryKey: keys.gifts.list({}),
		queryFn: () => listGifts({}),
	})

	if (isPending) return <p className="text-sm font-base text-foreground/60">Loading gifts…</p>
	if (isError) return <p className="text-sm font-base text-destructive">Failed to load gifts.</p>

	return (
		<div className="space-y-4">
			<div className="flex items-center justify-between">
				<h1 className="text-2xl font-heading">Gifts</h1>
				<Button asChild>
					<Link to="/gifts/new">New Gift</Link>
				</Button>
			</div>
			<GiftsTable
				data={data?.items ?? []}
			/>
		</div>
	)
}
