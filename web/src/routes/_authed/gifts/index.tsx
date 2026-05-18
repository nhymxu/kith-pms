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

	if (isPending) return <p className="text-[13px] text-zinc-500">Loading gifts…</p>
	if (isError) return <p className="text-[13px] text-red-600">Failed to load gifts.</p>

	return (
		<div className="space-y-4">
			<div className="flex items-center justify-between">
				<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">Gifts</h1>
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
