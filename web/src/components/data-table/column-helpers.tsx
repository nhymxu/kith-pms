import {
	type Column,
	type ColumnDef,
	columnFacetingFeature,
	columnFilteringFeature,
	columnSizingFeature,
	createFilteredRowModel,
	createPaginatedRowModel,
	createSortedRowModel,
	globalFilteringFeature,
	type RowData,
	rowPaginationFeature,
	rowSelectionFeature,
	rowSortingFeature,
	tableFeatures,
} from "@tanstack/react-table";
import type { ReactNode } from "react";

// The single v9 feature set shared by every table in the app: global search
// filtering, sorting, client-side pagination, row selection (the people
// table's bulk bar), column sizing (fixed-width display columns), and
// faceting (the filtered-row count in the pagination footer). The core row
// model is implicit in the table itself.
export const features = tableFeatures({
	columnFilteringFeature,
	columnFacetingFeature,
	columnSizingFeature,
	globalFilteringFeature,
	rowSortingFeature,
	rowPaginationFeature,
	rowSelectionFeature,
	filteredRowModel: createFilteredRowModel(),
	sortedRowModel: createSortedRowModel(),
	paginatedRowModel: createPaginatedRowModel(),
});

export type TableFeatures = typeof features;
export type TableColumn<T extends RowData> = ColumnDef<
	TableFeatures,
	T,
	unknown
>;
export type TableInstance<T extends RowData> = ReturnType<
	typeof import("@tanstack/react-table").useTable<TableFeatures, T>
>;
export type SortableColumn<T extends RowData> = Column<
	TableFeatures,
	T,
	unknown
>;

// Creates a sortable header cell — pass to `header` in your ColumnDef.
export function sortableHeader<T extends RowData>(label: string) {
	return ({ column }: { column: SortableColumn<T> }) => {
		const sorted = column.getIsSorted();
		const arrow = sorted === "asc" ? " ↑" : sorted === "desc" ? " ↓" : "";
		return (
			<button
				type="button"
				onClick={() => column.toggleSorting(sorted === "asc")}
				className="flex items-center gap-1 text-[11px] font-medium text-sub uppercase tracking-wider hover:text-ink transition-colors"
			>
				{label}
				<span className="text-accent-text">{arrow}</span>
			</button>
		);
	};
}

// Wraps a render fn into a ColumnDef `cell` that receives the row value.
export function valueCell<T extends RowData, V>(
	render: (value: V, row: T) => ReactNode,
): ColumnDef<TableFeatures, T, unknown>["cell"] {
	return ({ getValue, row }) => render(getValue() as V, row.original);
}

// A simple text column definition helper.
export function textColumn<T extends RowData>(
	id: keyof T & string,
	label: string,
	opts?: Partial<TableColumn<T>>,
): TableColumn<T> {
	return {
		id,
		accessorKey: id,
		header: sortableHeader<T>(label),
		enableSorting: true,
		...opts,
	};
}
