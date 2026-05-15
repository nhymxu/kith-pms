// Gifts table: columns for title, person, date, debt direction badge
import { useMemo } from "react"
import type { ColumnDef } from "@tanstack/react-table"
import { Link } from "@tanstack/react-router"
import { DataTable } from "#/components/data-table/data-table"
import { Badge } from "#/components/ui/badge"
import { sortableHeader, valueCell } from "#/components/data-table/column-helpers"
import type { GiftWithPerson } from "#/schemas/gift"

interface GiftsTableProps {
	data: GiftWithPerson[]
	toolbarActions?: React.ReactNode
}

function DebtBadge({ debtType, direction }: { debtType: string; direction: string }) {
	if (direction === "given") return <Badge variant="neutral">Given</Badge>
	if (direction === "received") return <Badge>Received</Badge>
	if (debtType === "i_owe")
		return <Badge className="bg-yellow-300 text-black border-black">I owe</Badge>
	if (debtType === "they_owe")
		return <Badge className="bg-green-300 text-black border-black">They owe</Badge>
	return <Badge variant="neutral">Planned</Badge>
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
						className="font-base underline hover:text-main"
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
				cell: valueCell<GiftWithPerson, string>((val) => formatDate(val)),
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
