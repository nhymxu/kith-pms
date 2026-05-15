import type { Table } from "@tanstack/react-table"
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from "lucide-react"
import { Button } from "#/components/ui/button"

interface DataTablePaginationProps<T> {
	table: Table<T>
}

export function DataTablePagination<T>({ table }: DataTablePaginationProps<T>) {
	return (
		<div className="flex items-center justify-between px-2 py-3 border-t-2 border-border">
			<p className="text-sm font-base text-foreground/70">
				{table.getFilteredSelectedRowModel().rows.length} of{" "}
				{table.getFilteredRowModel().rows.length} row(s) selected.
			</p>

			<div className="flex items-center gap-2">
				<p className="text-sm font-base">
					Page {table.getState().pagination.pageIndex + 1} of {table.getPageCount()}
				</p>
				<div className="flex items-center gap-1">
					<Button
						variant="neutral"
						size="icon"
						onClick={() => table.setPageIndex(0)}
						disabled={!table.getCanPreviousPage()}
						aria-label="First page"
					>
						<ChevronsLeft className="size-4" />
					</Button>
					<Button
						variant="neutral"
						size="icon"
						onClick={() => table.previousPage()}
						disabled={!table.getCanPreviousPage()}
						aria-label="Previous page"
					>
						<ChevronLeft className="size-4" />
					</Button>
					<Button
						variant="neutral"
						size="icon"
						onClick={() => table.nextPage()}
						disabled={!table.getCanNextPage()}
						aria-label="Next page"
					>
						<ChevronRight className="size-4" />
					</Button>
					<Button
						variant="neutral"
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
