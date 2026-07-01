import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { Button } from "#/components/ui/button";
import { listGifts } from "#/endpoints/gifts";
import { GiftsTable } from "#/features/gifts/gifts-table";
import { keys } from "#/query-keys";

export const Route = createFileRoute("/_authed/gifts/")({
	component: GiftsPage,
	pendingComponent: () => (
		<p className="text-[13px] text-zinc-500">Loading gifts…</p>
	),
	errorComponent: () => (
		<p className="text-[13px] text-red-600">Failed to load gifts.</p>
	),
});

function GiftsPage() {
	const { data } = useSuspenseQuery({
		queryKey: keys.gifts.list({}),
		queryFn: () => listGifts({}),
	});

	return (
		<div className="space-y-4">
			<div className="flex items-center justify-between">
				<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">
					Gifts
				</h1>
				<Button asChild>
					<Link to="/gifts/new">New Gift</Link>
				</Button>
			</div>
			<GiftsTable data={data.items} />
		</div>
	);
}
