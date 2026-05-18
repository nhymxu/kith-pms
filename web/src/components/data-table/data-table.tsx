import {
	flexRender,
	getCoreRowModel,
	getFilteredRowModel,
	getPaginationRowModel,
	getSortedRowModel,
	useReactTable,
	type ColumnDef,
	type SortingState,
} from "@tanstack/react-table"
import { useState, type ReactNode } from "react"
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "#/components/ui/table"
import { DataTableToolbar } from "./data-table-toolbar"
import { DataTablePagination } from "./data-table-pagination"

interface DataTableProps<T> {
	columns: ColumnDef<T>[]
	data: T[]
	pageSize?: number
	emptyState?: ReactNode
	toolbarActions?: ReactNode
	rowClassName?: (row: { original: T }) => string
}

export function DataTable<T>({
	columns,
	data,
	pageSize = 20,
	emptyState,
	toolbarActions,
	rowClassName,
}: DataTableProps<T>) {
	const [sorting, setSorting] = useState<SortingState>([])
	const [globalFilter, setGlobalFilter] = useState("")

	const table = useReactTable({
		data,
		columns,
		state: { sorting, globalFilter },
		onSortingChange: setSorting,
		onGlobalFilterChange: setGlobalFilter,
		getCoreRowModel: getCoreRowModel(),
		getSortedRowModel: getSortedRowModel(),
		getFilteredRowModel: getFilteredRowModel(),
		getPaginationRowModel: getPaginationRowModel(),
		initialState: { pagination: { pageSize } },
	})

	const rows = table.getRowModel().rows

	return (
		<div className="border border-zinc-200 rounded-md bg-white">
			<div className="px-4">
				<DataTableToolbar
					table={table}
					globalFilter={globalFilter}
					onGlobalFilterChange={setGlobalFilter}
				>
					{toolbarActions}
				</DataTableToolbar>
			</div>

			<Table>
				<TableHeader>
					{table.getHeaderGroups().map((hg) => (
						<TableRow key={hg.id}>
							{hg.headers.map((header) => (
								<TableHead key={header.id} style={{ width: header.getSize() }}>
									{header.isPlaceholder
										? null
										: flexRender(header.column.columnDef.header, header.getContext())}
								</TableHead>
							))}
						</TableRow>
					))}
				</TableHeader>

				<TableBody>
					{rows.length > 0 ? (
						rows.map((row) => (
							<TableRow key={row.id} data-state={row.getIsSelected() ? "selected" : undefined} className={rowClassName?.(row) ?? ""}>
								{row.getVisibleCells().map((cell) => (
									<TableCell key={cell.id}>
										{flexRender(cell.column.columnDef.cell, cell.getContext())}
									</TableCell>
								))}
							</TableRow>
						))
					) : (
						<TableRow>
							<TableCell colSpan={columns.length} className="h-32 text-center">
								{emptyState ?? (
									<span className="text-sm text-foreground/50">No results found.</span>
								)}
							</TableCell>
						</TableRow>
					)}
				</TableBody>
			</Table>

			<DataTablePagination table={table} />
		</div>
	)
}
