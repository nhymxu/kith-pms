import type { Table } from "@tanstack/react-table"
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from "lucide-react"
import { Button } from "#/components/ui/button"

interface DataTablePaginationProps<T> {
	table: Table<T>
}

export function DataTablePagination<T>({ table }: DataTablePaginationProps<T>) {
	return (
		<div className="flex items-center justify-between px-4 py-3 border-t border-zinc-200">
			<p className="font-mono text-[12px] text-zinc-500">
				{table.getFilteredSelectedRowModel().rows.length} of{" "}
				{table.getFilteredRowModel().rows.length} rows
			</p>

			<div className="flex items-center gap-2">
				<p className="font-mono text-[12px] text-zinc-500">
					Page {table.getState().pagination.pageIndex + 1} of {table.getPageCount()}
				</p>
				<div className="flex items-center gap-1">
					<Button
						variant="ghost"
						size="icon"
						onClick={() => table.setPageIndex(0)}
						disabled={!table.getCanPreviousPage()}
						aria-label="First page"
					>
						<ChevronsLeft className="size-4" />
					</Button>
					<Button
						variant="ghost"
						size="icon"
						onClick={() => table.previousPage()}
						disabled={!table.getCanPreviousPage()}
						aria-label="Previous page"
					>
						<ChevronLeft className="size-4" />
					</Button>
					<Button
						variant="ghost"
						size="icon"
						onClick={() => table.nextPage()}
						disabled={!table.getCanNextPage()}
						aria-label="Next page"
					>
						<ChevronRight className="size-4" />
					</Button>
					<Button
						variant="ghost"
						size="icon"
						onClick={() => table.setPageIndex(table.getPageCount() - 1)}
						disabled={!table.getCanNextPage()}
						aria-label="Last page"
					>
						<ChevronsRight className="size-4" />
					</Button>
				</div>
			</div>
		</div>
	)
}
