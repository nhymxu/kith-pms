import {
	flexRender,
	type OnChangeFn,
	type RowData,
	type RowSelectionState,
	type SortingState,
	useTable,
} from "@tanstack/react-table";
import { type ReactNode, useEffect, useState } from "react";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "#/components/ui/table";
import { features, type TableColumn } from "./column-helpers";
import { DataTablePagination } from "./data-table-pagination";
import { DataTableToolbar } from "./data-table-toolbar";

interface DataTableProps<T extends RowData> {
	columns: TableColumn<T>[];
	data: T[];
	pageSize?: number;
	emptyState?: ReactNode;
	toolbarActions?: ReactNode;
	rowClassName?: (row: { original: T }) => string;
	// Server-side pagination — provide all three to enable manual mode
	totalCount?: number;
	pageIndex?: number;
	onPageChange?: (pageIndex: number) => void;
	// Row selection — provide all three to enable
	enableRowSelection?: boolean;
	rowSelection?: RowSelectionState;
	onRowSelectionChange?: OnChangeFn<RowSelectionState>;
	getRowId?: (row: T) => string;
	hideToolbar?: boolean;
	pageSizeSelector?: ReactNode;
}

export function DataTable<T extends RowData>({
	columns,
	data,
	pageSize = 20,
	emptyState,
	toolbarActions,
	rowClassName,
	totalCount,
	pageIndex,
	onPageChange,
	enableRowSelection = false,
	rowSelection,
	onRowSelectionChange,
	getRowId,
	hideToolbar = false,
	pageSizeSelector,
}: DataTableProps<T>) {
	const [sorting, setSorting] = useState<SortingState>([]);
	const [globalFilter, setGlobalFilter] = useState("");
	const [internalPageIndex, setInternalPageIndex] = useState(0);

	const isServerPaginated =
		totalCount !== undefined &&
		pageIndex !== undefined &&
		onPageChange !== undefined;

	// Client-side pagination page count depends on `data.length / pageSize`;
	// reset to page 0 whenever either changes so the page index can't point
	// past the end (e.g. after picking a smaller page size or a new fetch).
	// biome-ignore lint/correctness/useExhaustiveDependencies: data/pageSize are props, not derived state — both must trigger the reset
	useEffect(() => {
		if (!isServerPaginated) setInternalPageIndex(0);
	}, [data, pageSize, isServerPaginated]);

	const resolvedPageIndex = isServerPaginated ? pageIndex : internalPageIndex;

	const checkboxCol: TableColumn<T> = {
		id: "select",
		size: 40,
		header: ({ table }) => (
			<input
				type="checkbox"
				className="size-4 cursor-pointer accent-accent"
				checked={table.getIsAllPageRowsSelected()}
				onChange={(e) => table.toggleAllPageRowsSelected(e.target.checked)}
				aria-label="Select all"
			/>
		),
		cell: ({ row }) => (
			<input
				type="checkbox"
				className="size-4 cursor-pointer accent-accent"
				checked={row.getIsSelected()}
				onChange={(e) => row.toggleSelected(e.target.checked)}
				onClick={(e) => e.stopPropagation()}
				aria-label="Select row"
			/>
		),
	};

	const resolvedCols = enableRowSelection ? [checkboxCol, ...columns] : columns;

	const table = useTable({
		features,
		data,
		columns: resolvedCols,
		state: {
			sorting,
			globalFilter,
			pagination: { pageIndex: resolvedPageIndex, pageSize },
			...(enableRowSelection && { rowSelection: rowSelection ?? {} }),
		},
		onSortingChange: setSorting,
		onGlobalFilterChange: setGlobalFilter,
		onPaginationChange: (updater) => {
			const next =
				typeof updater === "function"
					? updater({ pageIndex: resolvedPageIndex, pageSize })
					: updater;
			if (isServerPaginated) {
				onPageChange(next.pageIndex);
			} else {
				setInternalPageIndex(next.pageIndex);
			}
		},
		...(enableRowSelection && { onRowSelectionChange }),
		...(getRowId && { getRowId }),
		...(isServerPaginated && {
			manualPagination: true,
			rowCount: totalCount,
		}),
		enableRowSelection,
	});

	const rows = table.getRowModel().rows;

	return (
		<div className="border-bw border-line rounded-base bg-panel">
			{!hideToolbar && (
				<div className="px-4">
					<DataTableToolbar
						table={table}
						globalFilter={globalFilter}
						onGlobalFilterChange={setGlobalFilter}
					>
						{toolbarActions}
					</DataTableToolbar>
				</div>
			)}

			<Table className="table-fixed">
				<colgroup>
					{table.getHeaderGroups()[0]?.headers.map((header) => (
						<col
							key={header.id}
							style={
								header.column.columnDef.size
									? { width: header.getSize() }
									: undefined
							}
						/>
					))}
				</colgroup>
				<TableHeader>
					{table.getHeaderGroups().map((hg) => (
						<TableRow key={hg.id}>
							{hg.headers.map((header) => (
								<TableHead key={header.id} style={{ width: header.getSize() }}>
									{header.isPlaceholder
										? null
										: flexRender(
												header.column.columnDef.header,
												header.getContext(),
											)}
								</TableHead>
							))}
						</TableRow>
					))}
				</TableHeader>

				<TableBody>
					{rows.length > 0 ? (
						rows.map((row) => (
							<TableRow
								key={row.id}
								data-state={row.getIsSelected() ? "selected" : undefined}
								className={rowClassName?.(row) ?? ""}
							>
								{row.getAllCells().map((cell) => (
									<TableCell
										key={cell.id}
										style={
											cell.column.columnDef.size
												? { width: cell.column.getSize() }
												: undefined
										}
									>
										{flexRender(cell.column.columnDef.cell, cell.getContext())}
									</TableCell>
								))}
							</TableRow>
						))
					) : (
						<TableRow>
							<TableCell
								colSpan={resolvedCols.length}
								className="h-32 text-center"
							>
								{emptyState ?? (
									<span className="text-sm text-foreground/50">
										No results found.
									</span>
								)}
							</TableCell>
						</TableRow>
					)}
				</TableBody>
			</Table>

			<DataTablePagination table={table} pageSizeSelector={pageSizeSelector} />
		</div>
	);
}
