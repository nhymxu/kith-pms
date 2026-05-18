import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { useQuery } from "@tanstack/react-query"
import { z } from "zod"
import { listAudit } from "#/endpoints/audit"
import { keys } from "#/query-keys"
import { AuditTable } from "#/features/audit/audit-table"
import { Button } from "#/components/ui/button"

const searchSchema = z.object({
	page: z.coerce.number().min(1).optional().default(1),
	from_date: z.string().optional(),
	to_date: z.string().optional(),
})

export const Route = createFileRoute("/_authed/audit/")({
	validateSearch: searchSchema,
	component: AuditPage,
})

function AuditPage() {
	const navigate = useNavigate()
	const search = Route.useSearch()

	const { data, isPending, isError } = useQuery({
		queryKey: keys.audit.list({ page: search.page }),
		queryFn: () => listAudit({ page: search.page }),
	})

	if (isError) return <p className="text-[13px] text-red-600">Failed to load audit log.</p>

	return (
		<div className="space-y-4">
			<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">Audit Log</h1>

			<div className="flex flex-wrap gap-3 items-end">
				<div className="space-y-1">
					<label className="text-[11px] font-medium text-zinc-500">From</label>
					<input
						type="date"
						value={search.from_date ?? ""}
						onChange={(e) =>
							void navigate({ to: "/audit", search: { ...search, from_date: e.target.value || undefined, page: 1 } })
						}
						className="h-9 border border-zinc-200 rounded-md bg-white px-3 text-[13px] focus:outline-none focus:ring-2 focus:ring-indigo-600"
					/>
				</div>
				<div className="space-y-1">
					<label className="text-[11px] font-medium text-zinc-500">To</label>
					<input
						type="date"
						value={search.to_date ?? ""}
						onChange={(e) =>
							void navigate({ to: "/audit", search: { ...search, to_date: e.target.value || undefined, page: 1 } })
						}
						className="h-9 border border-zinc-200 rounded-md bg-white px-3 text-[13px] focus:outline-none focus:ring-2 focus:ring-indigo-600"
					/>
				</div>
				{(search.from_date || search.to_date) && (
					<Button
						variant="neutral"
						size="sm"
						onClick={() =>
							void navigate({ to: "/audit", search: { ...search, from_date: undefined, to_date: undefined, page: 1 } })
						}
					>
						Clear dates
					</Button>
				)}
			</div>

			{isPending ? (
				<p className="text-[13px] text-zinc-500 py-4">Loading…</p>
			) : (
				<>
					<AuditTable data={data?.data ?? []} />
					{data?.has_more && (
						<Button
							variant="neutral"
							size="sm"
							onClick={() => void navigate({ to: "/audit", search: { ...search, page: (search.page ?? 1) + 1 } })}
						>
							Load more
						</Button>
					)}
				</>
			)}
		</div>
	)
}
