import {
	type ColumnDef,
	flexRender,
	getCoreRowModel,
	getFilteredRowModel,
	getPaginationRowModel,
	getSortedRowModel,
	type OnChangeFn,
	type RowSelectionState,
	type SortingState,
	useReactTable,
} from "@tanstack/react-table";
import { type ReactNode, useState } from "react";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "#/components/ui/table";
import { DataTablePagination } from "./data-table-pagination";
import { DataTableToolbar } from "./data-table-toolbar";

interface DataTableProps<T> {
	columns: ColumnDef<T>[];
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
}

export function DataTable<T>({
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
}: DataTableProps<T>) {
	const [sorting, setSorting] = useState<SortingState>([]);
	const [globalFilter, setGlobalFilter] = useState("");

	const isServerPaginated =
		totalCount !== undefined &&
		pageIndex !== undefined &&
		onPageChange !== undefined;

	const checkboxCol: ColumnDef<T> = {
		id: "select",
		size: 40,
		header: ({ table }) => (
			<input
				type="checkbox"
				className="size-4 cursor-pointer accent-indigo-600"
				checked={table.getIsAllPageRowsSelected()}
				onChange={(e) => table.toggleAllPageRowsSelected(e.target.checked)}
				aria-label="Select all"
			/>
		),
		cell: ({ row }) => (
			<input
				type="checkbox"
				className="size-4 cursor-pointer accent-indigo-600"
				checked={row.getIsSelected()}
				onChange={(e) => row.toggleSelected(e.target.checked)}
				onClick={(e) => e.stopPropagation()}
				aria-label="Select row"
			/>
		),
	};

	const resolvedCols = enableRowSelection ? [checkboxCol, ...columns] : columns;

	const table = useReactTable({
		data,
		columns: resolvedCols,
		state: {
			sorting,
			globalFilter,
			...(isServerPaginated && {
				pagination: { pageIndex, pageSize },
			}),
			...(enableRowSelection && { rowSelection: rowSelection ?? {} }),
		},
		onSortingChange: setSorting,
		onGlobalFilterChange: setGlobalFilter,
		...(enableRowSelection && { onRowSelectionChange }),
		...(getRowId && { getRowId }),
		...(isServerPaginated && {
			manualPagination: true,
			rowCount: totalCount,
			onPaginationChange: (updater) => {
				const next =
					typeof updater === "function"
						? updater({ pageIndex, pageSize })
						: updater;
				onPageChange(next.pageIndex);
			},
		}),
		enableRowSelection,
		getCoreRowModel: getCoreRowModel(),
		getSortedRowModel: getSortedRowModel(),
		getFilteredRowModel: getFilteredRowModel(),
		getPaginationRowModel: getPaginationRowModel(),
		initialState: { pagination: { pageSize } },
	});

	const rows = table.getRowModel().rows;

	return (
		<div className="border border-zinc-200 rounded-md bg-white">
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
								{row.getVisibleCells().map((cell) => (
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

			<DataTablePagination table={table} />
		</div>
	);
}
