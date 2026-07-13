import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { z } from "zod";
import { PageSizeSelector } from "#/components/page-size-selector";
import { Button } from "#/components/ui/button";
import { listGifts } from "#/endpoints/gifts";
import { getSettings } from "#/endpoints/settings";
import { GiftsTable } from "#/features/gifts/gifts-table";
import { usePageSizeOverride } from "#/lib/use-page-size-override";
import { keys } from "#/query-keys";

const searchSchema = z.object({
	page: z.coerce.number().min(1).optional().default(1),
	page_size: z.coerce.number().min(1).max(200).optional(),
});

export const Route = createFileRoute("/_authed/gifts/")({
	validateSearch: searchSchema,
	component: GiftsPage,
});

function GiftsPage() {
	const navigate = useNavigate();
	const search = Route.useSearch();

	const { data: settingsData } = useQuery({
		queryKey: ["settings"],
		queryFn: getSettings,
	});
	const { override, setPageSize, clear } = usePageSizeOverride("gifts");
	const effectivePageSize =
		search.page_size ?? override ?? settingsData?.default_page_size ?? 20;

	const { data } = useQuery({
		queryKey: keys.gifts.list({
			page: search.page,
			page_size: effectivePageSize,
		}),
		queryFn: () =>
			listGifts({ page: search.page, page_size: effectivePageSize }),
		placeholderData: keepPreviousData,
	});

	return (
		<div className="space-y-4">
			<div className="flex items-center justify-between">
				<h1 className="text-[18px] font-semibold tracking-tight text-ink font-display">
					Gifts
				</h1>
				<Button asChild>
					<Link to="/gifts/new">New Gift</Link>
				</Button>
			</div>
			<GiftsTable
				data={data?.items ?? []}
				pageSize={effectivePageSize}
				totalCount={data?.total ?? 0}
				pageIndex={search.page - 1}
				onPageChange={(idx) =>
					void navigate({ to: "/gifts", search: { ...search, page: idx + 1 } })
				}
				pageSizeSelector={
					<PageSizeSelector
						value={effectivePageSize}
						hasOverride={override !== null}
						onChange={(n) => {
							setPageSize(n);
							void navigate({
								to: "/gifts",
								search: { ...search, page_size: undefined, page: 1 },
							});
						}}
						onReset={() => {
							clear();
							void navigate({
								to: "/gifts",
								search: { ...search, page_size: undefined, page: 1 },
							});
						}}
					/>
				}
			/>
		</div>
	);
}
