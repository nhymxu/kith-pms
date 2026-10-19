import type { RowData } from "@tanstack/react-table";
import {
	ChevronLeft,
	ChevronRight,
	ChevronsLeft,
	ChevronsRight,
} from "lucide-react";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "#/components/ui/button";
import type { TableInstance } from "./column-helpers";

interface DataTablePaginationProps<T extends RowData> {
	table: TableInstance<T>;
	pageSizeSelector?: ReactNode;
}

export function DataTablePagination<T extends RowData>({
	table,
	pageSizeSelector,
}: DataTablePaginationProps<T>) {
	const { t } = useTranslation();
	const shown = table.options.manualPagination
		? table.getRowModel().rows.length
		: table.getFilteredSelectedRowModel().rows.length;
	const total = table.options.manualPagination
		? (table.options.rowCount ?? table.getFilteredRowModel().rows.length)
		: table.getFilteredRowModel().rows.length;
	return (
		<div className="flex items-center justify-between px-4 py-3 border-t border-line">
			<div className="flex items-center gap-4">
				<p className="font-mono text-[12px] text-sub">
					{t("table.rowsOf", { shown, total })}
				</p>
				{pageSizeSelector}
			</div>

			<div className="flex items-center gap-2">
				<p className="font-mono text-[12px] text-sub">
					{t("table.pageOf", {
						page: table.state.pagination.pageIndex + 1,
						pages: table.getPageCount(),
					})}
				</p>
				<div className="flex items-center gap-1">
					<Button
						variant="ghost"
						size="icon"
						onClick={() => table.setPageIndex(0)}
						disabled={!table.getCanPreviousPage()}
						aria-label={t("table.firstPage")}
					>
						<ChevronsLeft className="size-4" />
					</Button>
					<Button
						variant="ghost"
						size="icon"
						onClick={() => table.previousPage()}
						disabled={!table.getCanPreviousPage()}
						aria-label={t("table.previousPage")}
					>
						<ChevronLeft className="size-4" />
					</Button>
					<Button
						variant="ghost"
						size="icon"
						onClick={() => table.nextPage()}
						disabled={!table.getCanNextPage()}
						aria-label={t("table.nextPage")}
					>
						<ChevronRight className="size-4" />
					</Button>
					<Button
						variant="ghost"
						size="icon"
						onClick={() => table.setPageIndex(table.getPageCount() - 1)}
						disabled={!table.getCanNextPage()}
						aria-label={t("table.lastPage")}
					>
						<ChevronsRight className="size-4" />
					</Button>
				</div>
			</div>
		</div>
	);
}
