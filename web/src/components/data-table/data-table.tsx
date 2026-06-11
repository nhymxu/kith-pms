import {
	type ColumnDef,
	flexRender,
	getCoreRowModel,
	getFilteredRowModel,
	getPaginationRowModel,
	getSortedRowModel,
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
}: DataTableProps<T>) {
	const [sorting, setSorting] = useState<SortingState>([]);
	const [globalFilter, setGlobalFilter] = useState("");

	const isServerPaginated =
		totalCount !== undefined &&
		pageIndex !== undefined &&
		onPageChange !== undefined;

	const table = useReactTable({
		data,
		columns,
		state: {
			sorting,
			globalFilter,
			...(isServerPaginated && {
				pagination: { pageIndex, pageSize },
			}),
		},
		onSortingChange: setSorting,
		onGlobalFilterChange: setGlobalFilter,
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
		getCoreRowModel: getCoreRowModel(),
		getSortedRowModel: getSortedRowModel(),
		getFilteredRowModel: getFilteredRowModel(),
		getPaginationRowModel: getPaginationRowModel(),
		initialState: { pagination: { pageSize } },
	});

	const rows = table.getRowModel().rows;

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
