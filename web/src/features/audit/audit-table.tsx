// Audit log table: read-only, columns for timestamp, actor, action, target, detail

import type { ColumnDef } from "@tanstack/react-table";
import { useMemo } from "react";
import {
	sortableHeader,
	valueCell,
} from "#/components/data-table/column-helpers";
import { DataTable } from "#/components/data-table/data-table";
import { Pill, type PillVariant } from "#/components/ui/pill";
import { formatDateTime } from "#/lib/format-datetime";
import type { AuditEntry, AuditMetadata } from "#/schemas/audit";

interface AuditTableProps {
	data: AuditEntry[];
	toolbarActions?: React.ReactNode;
	pageSizeSelector?: React.ReactNode;
	pageSize?: number;
	totalCount?: number;
	pageIndex?: number;
	onPageChange?: (pageIndex: number) => void;
}

const ACTION_VARIANTS: Record<string, PillVariant> = {
	create: "success",
	update: "accent",
	delete: "danger",
	login: "plain",
	logout: "plain",
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
	if (!meta) return <span className="text-sub text-[12px]">—</span>;

	const label = meta.detail_action
		? (DETAIL_ACTION_LABELS[meta.detail_action] ??
			meta.detail_action.replace(/_/g, " "))
		: null;

	return (
		<div className="space-y-1">
			{label && <Pill variant="plain">{label}</Pill>}
			{meta.label && <span className="text-[13px] text-ink">{meta.label}</span>}
			{meta.changes && meta.changes.length > 0 && (
				<div className="space-y-0.5">
					{meta.changes.map((c) => (
						<div
							key={c.field}
							className="flex items-baseline gap-1 text-[11px] leading-tight"
						>
							<span className="text-sub font-mono shrink-0">{c.field}:</span>
							<span
								className="text-danger-fg line-through truncate max-w-[80px]"
								title={formatValue(c.old)}
							>
								{formatValue(c.old)}
							</span>
							<span className="text-sub">→</span>
							<span
								className="text-success-fg truncate max-w-[80px]"
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

export function AuditTable({
	data,
	toolbarActions,
	pageSizeSelector,
	pageSize,
	totalCount,
	pageIndex,
	onPageChange,
}: AuditTableProps) {
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
							<span className="font-mono text-[12px] text-sub">
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
					<Pill variant={ACTION_VARIANTS[val] ?? "plain"}>{val}</Pill>
				)),
			},
			{
				id: "entity_type",
				accessorKey: "entity_type",
				header: sortableHeader<AuditEntry>("Type"),
				enableSorting: true,
				cell: valueCell<AuditEntry, string>((val) => (
					<span className="text-[12px] text-sub capitalize">
						{val.replace("_", " ")}
					</span>
				)),
			},
			{
				id: "entity_name",
				accessorKey: "entity_name",
				header: "Target",
				cell: valueCell<AuditEntry, string>((val) => (
					<span className="text-[13px] text-ink">{val || "—"}</span>
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
			pageSizeSelector={pageSizeSelector}
			pageSize={pageSize}
			totalCount={totalCount}
			pageIndex={pageIndex}
			onPageChange={onPageChange}
			emptyState={
				<span className="text-sm text-foreground/50">No audit entries.</span>
			}
		/>
	);
}
