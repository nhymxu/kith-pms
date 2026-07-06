import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { z } from "zod";
import { PageSizeSelector } from "#/components/page-size-selector";
import { Button } from "#/components/ui/button";
import { listAudit } from "#/endpoints/audit";
import { getSettings } from "#/endpoints/settings";
import { AuditTable } from "#/features/audit/audit-table";
import { usePageSizeOverride } from "#/lib/use-page-size-override";
import { keys } from "#/query-keys";

const searchSchema = z.object({
	from_date: z.string().optional(),
	to_date: z.string().optional(),
	page: z.coerce.number().min(1).optional().default(1),
	page_size: z.coerce.number().min(1).max(200).optional(),
});

export const Route = createFileRoute("/_authed/audit/")({
	validateSearch: searchSchema,
	component: AuditPage,
});

function AuditPage() {
	const navigate = useNavigate();
	const search = Route.useSearch();

	const { data: settingsData } = useQuery({
		queryKey: ["settings"],
		queryFn: getSettings,
	});
	const { override, setPageSize, clear } = usePageSizeOverride("audit");
	const effectivePageSize =
		search.page_size ?? override ?? settingsData?.default_page_size ?? 20;

	const { data } = useQuery({
		queryKey: keys.audit.list({
			page: search.page,
			from_date: search.from_date,
			to_date: search.to_date,
			page_size: effectivePageSize,
		}),
		queryFn: () =>
			listAudit({
				page: search.page,
				from_date: search.from_date,
				to_date: search.to_date,
				page_size: effectivePageSize,
			}),
		placeholderData: keepPreviousData,
	});

	return (
		<div className="space-y-4">
			<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">
				Audit Log
			</h1>

			<div className="flex flex-wrap gap-3 items-end">
				<div className="space-y-1">
					<label
						htmlFor="from-date"
						className="text-[11px] font-medium text-zinc-500"
					>
						From
					</label>
					<input
						id="from-date"
						type="date"
						value={search.from_date ?? ""}
						onChange={(e) =>
							void navigate({
								to: "/audit",
								search: {
									...search,
									from_date: e.target.value || undefined,
									page: 1,
								},
							})
						}
						className="h-9 border border-zinc-200 rounded-md bg-white px-3 text-[13px] focus:outline-none focus:ring-2 focus:ring-indigo-600"
					/>
				</div>
				<div className="space-y-1">
					<label
						htmlFor="to-date"
						className="text-[11px] font-medium text-zinc-500"
					>
						To
					</label>
					<input
						id="to-date"
						type="date"
						value={search.to_date ?? ""}
						onChange={(e) =>
							void navigate({
								to: "/audit",
								search: {
									...search,
									to_date: e.target.value || undefined,
									page: 1,
								},
							})
						}
						className="h-9 border border-zinc-200 rounded-md bg-white px-3 text-[13px] focus:outline-none focus:ring-2 focus:ring-indigo-600"
					/>
				</div>
				{(search.from_date || search.to_date) && (
					<Button
						variant="neutral"
						size="sm"
						onClick={() =>
							void navigate({
								to: "/audit",
								search: {
									...search,
									from_date: undefined,
									to_date: undefined,
									page: 1,
								},
							})
						}
					>
						Clear dates
					</Button>
				)}
			</div>

			<AuditTable
				data={data?.data ?? []}
				pageSize={effectivePageSize}
				totalCount={data?.total ?? 0}
				pageIndex={search.page - 1}
				onPageChange={(idx) =>
					void navigate({ to: "/audit", search: { ...search, page: idx + 1 } })
				}
				pageSizeSelector={
					<PageSizeSelector
						value={effectivePageSize}
						hasOverride={override !== null}
						onChange={(n) => {
							setPageSize(n);
							void navigate({
								to: "/audit",
								search: { ...search, page_size: undefined, page: 1 },
							});
						}}
						onReset={() => {
							clear();
							void navigate({
								to: "/audit",
								search: { ...search, page_size: undefined, page: 1 },
							});
						}}
					/>
				}
			/>
		</div>
	);
}
