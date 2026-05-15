import type { Column, ColumnDef } from "@tanstack/react-table"
import type { ReactNode } from "react"

// Creates a sortable header cell — pass to `header` in your ColumnDef.
export function sortableHeader<T>(label: string) {
	return ({ column }: { column: Column<T, unknown> }) => {
		const sorted = column.getIsSorted()
		const arrow = sorted === "asc" ? " ↑" : sorted === "desc" ? " ↓" : ""
		return (
			<button
				type="button"
				onClick={() => column.toggleSorting(sorted === "asc")}
				className="flex items-center gap-1 font-heading text-foreground hover:text-main transition-colors"
			>
				{label}
				<span className="text-main">{arrow}</span>
			</button>
		)
	}
}

// Wraps a render fn into a ColumnDef `cell` that receives the row value.
export function valueCell<T, V>(render: (value: V, row: T) => ReactNode): ColumnDef<T>["cell"] {
	return ({ getValue, row }) => render(getValue() as V, row.original)
}

// A simple text column definition helper.
export function textColumn<T>(
	id: keyof T & string,
	label: string,
	opts?: Partial<ColumnDef<T>>,
): ColumnDef<T> {
	return {
		id,
		accessorKey: id,
		header: sortableHeader<T>(label),
		enableSorting: true,
		...opts,
	}
}
