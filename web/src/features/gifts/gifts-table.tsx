// Gifts table: columns for title, person, date, debt direction badge
import { useMemo } from "react"
import type { ColumnDef } from "@tanstack/react-table"
import { Link } from "@tanstack/react-router"
import { DataTable } from "#/components/data-table/data-table"
import { sortableHeader, valueCell } from "#/components/data-table/column-helpers"
import type { GiftWithPerson } from "#/schemas/gift"

interface GiftsTableProps {
	data: GiftWithPerson[]
	toolbarActions?: React.ReactNode
}

function DebtBadge({ debtType, direction }: { debtType: string; direction: string }) {
	if (direction === "given") return <span className="font-mono text-[10px] uppercase text-zinc-500">Given</span>
	if (direction === "received") return <span className="font-mono text-[10px] uppercase text-indigo-600">Received</span>
	if (debtType === "i_owe") return <span className="font-mono text-[10px] uppercase text-amber-600">I owe</span>
	if (debtType === "they_owe") return <span className="font-mono text-[10px] uppercase text-emerald-600">They owe</span>
	return <span className="font-mono text-[10px] uppercase text-zinc-400">Planned</span>
}

function formatDate(dateStr: string) {
	if (!dateStr) return "—"
	try {
		return new Date(dateStr).toLocaleDateString()
	} catch {
		return dateStr
	}
}

export function GiftsTable({ data, toolbarActions }: GiftsTableProps) {
	const columns = useMemo<ColumnDef<GiftWithPerson>[]>(
		() => [
			{
				id: "title",
				accessorKey: "title",
				header: sortableHeader<GiftWithPerson>("Gift"),
				enableSorting: true,
				cell: valueCell<GiftWithPerson, string>((val, row) => (
					<Link
						to="/gifts/$giftId"
						params={{ giftId: String(row.id) }}
						className="text-[13px] text-zinc-900 hover:text-indigo-600 hover:underline"
					>
						{val}
					</Link>
				)),
			},
			{
				id: "person_name",
				accessorKey: "person_name",
				header: sortableHeader<GiftWithPerson>("Person"),
				enableSorting: true,
			},
			{
				id: "date",
				accessorKey: "date",
				header: sortableHeader<GiftWithPerson>("Date"),
				enableSorting: true,
				cell: valueCell<GiftWithPerson, string>((val) =>
					val ? <span className="font-mono text-[12px] text-zinc-500">{formatDate(val)}</span> : <span className="text-zinc-300">—</span>
				),
			},
			{
				id: "debt",
				header: "Direction",
				cell: ({ row }) => (
					<DebtBadge debtType={row.original.debt_type ?? ""} direction={row.original.direction} />
				),
			},
		],
		[],
	)

	return (
		<DataTable
			columns={columns}
			data={data}
			toolbarActions={toolbarActions}
			emptyState={<span className="text-sm text-foreground/50">No gifts yet.</span>}
		/>
	)
}
