// Audit log table: read-only, columns for timestamp, actor, action, target, detail

import type { ColumnDef } from "@tanstack/react-table";
import { useMemo } from "react";
import {
	sortableHeader,
	valueCell,
} from "#/components/data-table/column-helpers";
import { DataTable } from "#/components/data-table/data-table";
import { formatDateTime } from "#/lib/format-datetime";
import type { AuditEntry, AuditMetadata } from "#/schemas/audit";

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

const DETAIL_ACTION_LABELS: Record<string, string> = {
	profile_update: "Profile",
	avatar_upload: "Avatar upload",
	avatar_delete: "Avatar delete",
	set_self: "Set self",
	last_contact_update: "Last contact",
};

function formatValue(val: unknown): string {
	if (val === null || val === undefined || val === "") return "—";
	return String(val);
}

function MetadataCell({ meta }: { meta: AuditMetadata | null | undefined }) {
	if (!meta) return <span className="text-zinc-400 text-[12px]">—</span>;

	const label = meta.detail_action
		? (DETAIL_ACTION_LABELS[meta.detail_action] ??
			meta.detail_action.replace(/_/g, " "))
		: null;

	return (
		<div className="space-y-1">
			{label && (
				<span className="inline-block text-[11px] font-medium text-indigo-700 bg-indigo-50 border border-indigo-200 rounded px-1.5 py-0.5 capitalize">
					{label}
				</span>
			)}
			{meta.changes && meta.changes.length > 0 && (
				<div className="space-y-0.5">
					{meta.changes.map((c) => (
						<div
							key={c.field}
							className="flex items-baseline gap-1 text-[11px] leading-tight"
						>
							<span className="text-zinc-500 font-mono shrink-0">
								{c.field}:
							</span>
							<span
								className="text-red-500 line-through truncate max-w-[80px]"
								title={formatValue(c.old)}
							>
								{formatValue(c.old)}
							</span>
							<span className="text-zinc-400">→</span>
							<span
								className="text-emerald-700 truncate max-w-[80px]"
								title={formatValue(c.new)}
							>
								{formatValue(c.new)}
							</span>
						</div>
					))}
				</div>
			)}
		</div>
	);
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
			{
				id: "metadata",
				accessorKey: "metadata",
				header: "Detail",
				cell: ({ row }) => <MetadataCell meta={row.original.metadata} />,
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
