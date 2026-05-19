// Audit log table: read-only, columns for timestamp, actor, action, target

import type { ColumnDef } from "@tanstack/react-table";
import { useMemo } from "react";
import {
	sortableHeader,
	valueCell,
} from "#/components/data-table/column-helpers";
import { DataTable } from "#/components/data-table/data-table";
import { formatDateTime } from "#/lib/format-datetime";
import type { AuditEntry } from "#/schemas/audit";

interface AuditTableProps {
	data: AuditEntry[];
	toolbarActions?: React.ReactNode;
}

const ACTION_COLORS: Record<string, string> = {
	create: "text-emerald-700",
	update: "text-indigo-700",
	delete: "text-red-700",
	login: "text-zinc-600",
	logout: "text-zinc-600",
};

export function AuditTable({ data, toolbarActions }: AuditTableProps) {
	const columns = useMemo<ColumnDef<AuditEntry>[]>(
		() => [
			{
				id: "created_at",
				accessorKey: "created_at",
				header: sortableHeader<AuditEntry>("Timestamp"),
				enableSorting: true,
				cell: valueCell<AuditEntry, string>((val) => {
					if (!val) return "—";
					try {
						return (
							<span className="font-mono text-[12px] text-zinc-500">
								{formatDateTime(val)}
							</span>
						);
					} catch {
						return val;
					}
				}),
			},
			{
				id: "action",
				accessorKey: "action",
				header: sortableHeader<AuditEntry>("Action"),
				enableSorting: true,
				cell: valueCell<AuditEntry, string>((val) => (
					<span
						className={`font-mono text-[12px] uppercase ${ACTION_COLORS[val] ?? "text-zinc-600"}`}
					>
						{val}
					</span>
				)),
			},
			{
				id: "entity_type",
				accessorKey: "entity_type",
				header: sortableHeader<AuditEntry>("Type"),
				enableSorting: true,
				cell: valueCell<AuditEntry, string>((val) => (
					<span className="text-[12px] text-zinc-500 capitalize">
						{val.replace("_", " ")}
					</span>
				)),
			},
			{
				id: "entity_name",
				accessorKey: "entity_name",
				header: "Target",
				cell: valueCell<AuditEntry, string>((val) => (
					<span className="text-[13px] text-zinc-700">{val || "—"}</span>
				)),
			},
		],
		[],
	);

	return (
		<DataTable
			columns={columns}
			data={data}
			toolbarActions={toolbarActions}
			emptyState={
				<span className="text-sm text-foreground/50">No audit entries.</span>
			}
		/>
	);
}
