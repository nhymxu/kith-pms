// Journal table: columns for date, people chips, title/summary, actions
import { useMemo } from "react"
import type { ColumnDef } from "@tanstack/react-table"
import { Link } from "@tanstack/react-router"
import { DataTable } from "#/components/data-table/data-table"
import { Badge } from "#/components/ui/badge"
import { sortableHeader, valueCell } from "#/components/data-table/column-helpers"
import type { JournalActivity } from "#/schemas/journal"

interface JournalTableProps {
	data: JournalActivity[]
	toolbarActions?: React.ReactNode
}

export function JournalTable({ data, toolbarActions }: JournalTableProps) {
	const columns = useMemo<ColumnDef<JournalActivity>[]>(
		() => [
			{
				id: "occurred_at_date",
				accessorKey: "occurred_at_date",
				header: sortableHeader<JournalActivity>("Date"),
				enableSorting: true,
				cell: valueCell<JournalActivity, string>((val) =>
					val ? new Date(val).toLocaleDateString() : "—",
				),
			},
			{
				id: "title",
				accessorKey: "title",
				header: sortableHeader<JournalActivity>("Title"),
				enableSorting: true,
				cell: valueCell<JournalActivity, string>((val, row) => (
					<Link
						to="/journal/$entryId"
						params={{ entryId: String(row.id) }}
						className="font-base underline hover:text-main"
					>
						{val}
					</Link>
				)),
			},
			{
				id: "people",
				header: "People",
				cell: ({ row }) => {
					const people = row.original.people ?? []
					if (!people.length) return <span className="text-foreground/40 text-xs">—</span>
					return (
						<div className="flex flex-wrap gap-1">
							{people.slice(0, 3).map((p) => (
								<Badge key={p.person_id} variant="neutral" className="text-xs">
									{p.name}
								</Badge>
							))}
							{people.length > 3 && (
								<Badge variant="neutral" className="text-xs">+{people.length - 3}</Badge>
							)}
						</div>
					)
				},
			},
			{
				id: "content_preview",
				header: "Preview",
				cell: ({ row }) => {
					const content = row.original.content ?? ""
					const preview = content.length > 80 ? `${content.slice(0, 80)}…` : content
					return <span className="text-xs text-foreground/60 font-base">{preview || "—"}</span>
				},
			},
		],
		[],
	)

	return (
		<DataTable
			columns={columns}
			data={data}
			toolbarActions={toolbarActions}
			emptyState={<span className="text-sm text-foreground/50">No journal entries yet.</span>}
		/>
	)
}
