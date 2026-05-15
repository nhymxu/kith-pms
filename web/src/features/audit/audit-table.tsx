// Audit log table: read-only, columns for timestamp, actor, action, target
import { useMemo } from "react"
import type { ColumnDef } from "@tanstack/react-table"
import { DataTable } from "#/components/data-table/data-table"
import { Badge } from "#/components/ui/badge"
import { sortableHeader, valueCell } from "#/components/data-table/column-helpers"
import type { AuditEntry } from "#/schemas/audit"

interface AuditTableProps {
	data: AuditEntry[]
	toolbarActions?: React.ReactNode
}

const ACTION_COLORS: Record<string, string> = {
	create: "bg-green-300 text-black border-black",
	update: "bg-yellow-300 text-black border-black",
	delete: "bg-red-300 text-black border-black",
}

export function AuditTable({ data, toolbarActions }: AuditTableProps) {
	const columns = useMemo<ColumnDef<AuditEntry>[]>(
		() => [
			{
				id: "created_at",
				accessorKey: "created_at",
				header: sortableHeader<AuditEntry>("Timestamp"),
				enableSorting: true,
				cell: valueCell<AuditEntry, string>((val) => {
					if (!val) return "—"
					try {
						return new Date(val).toLocaleString()
					} catch {
						return val
					}
				}),
			},
			{
				id: "action",
				accessorKey: "action",
				header: sortableHeader<AuditEntry>("Action"),
				enableSorting: true,
				cell: valueCell<AuditEntry, string>((val) => (
					<Badge className={ACTION_COLORS[val] ?? ""}>{val}</Badge>
				)),
			},
			{
				id: "entity_type",
				accessorKey: "entity_type",
				header: sortableHeader<AuditEntry>("Type"),
				enableSorting: true,
				cell: valueCell<AuditEntry, string>((val) => (
					<span className="text-xs font-base capitalize">{val.replace("_", " ")}</span>
				)),
			},
			{
				id: "entity_name",
				accessorKey: "entity_name",
				header: "Target",
				cell: valueCell<AuditEntry, string>((val) => (
					<span className="text-sm font-base">{val || "—"}</span>
				)),
			},
		],
		[],
	)

	return (
		<DataTable
			columns={columns}
			data={data}
			toolbarActions={toolbarActions}
			emptyState={<span className="text-sm text-foreground/50">No audit entries.</span>}
		/>
	)
}
