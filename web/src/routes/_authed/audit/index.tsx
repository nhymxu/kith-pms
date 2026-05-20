import { useQuery } from "@tanstack/react-query";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { z } from "zod";
import { Button } from "#/components/ui/button";
import { listAudit } from "#/endpoints/audit";
import { AuditTable } from "#/features/audit/audit-table";
import { keys } from "#/query-keys";

const searchSchema = z.object({
	from_date: z.string().optional(),
	to_date: z.string().optional(),
});

export const Route = createFileRoute("/_authed/audit/")({
	validateSearch: searchSchema,
	component: AuditPage,
});

function AuditPage() {
	const navigate = useNavigate();
	const search = Route.useSearch();

	const { data, isPending, isError } = useQuery({
		queryKey: keys.audit.list({}),
		queryFn: () => listAudit({}),
	});

	if (isError)
		return (
			<p className="text-[13px] text-red-600">Failed to load audit log.</p>
		);

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
									from_date: undefined,
									to_date: undefined,
								},
							})
						}
					>
						Clear dates
					</Button>
				)}
			</div>

			{isPending ? (
				<p className="text-[13px] text-zinc-500 py-4">Loading…</p>
			) : (
				<AuditTable data={data?.data ?? []} />
			)}
		</div>
	);
}
