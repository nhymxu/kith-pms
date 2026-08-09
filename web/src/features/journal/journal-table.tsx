// Journal table: columns for date, people chips, title/summary, actions

import { Link } from "@tanstack/react-router";
import { useMemo } from "react";
import {
	sortableHeader,
	type TableColumn,
	valueCell,
} from "#/components/data-table/column-helpers";
import { DataTable } from "#/components/data-table/data-table";
import { PersonChip } from "#/features/journal/person-label-chip";
import { formatDate } from "#/lib/format-datetime";
import type { JournalActivity } from "#/schemas/journal";

interface JournalTableProps {
	data: JournalActivity[];
	toolbarActions?: React.ReactNode;
}

export function JournalTable({ data, toolbarActions }: JournalTableProps) {
	const columns = useMemo<TableColumn<JournalActivity>[]>(
		() => [
			{
				id: "occurred_at_date",
				accessorKey: "occurred_at_date",
				header: sortableHeader<JournalActivity>("Date"),
				enableSorting: true,
				cell: valueCell<JournalActivity, string>((val) =>
					val ? (
						<span className="font-mono text-[12px] text-sub">
							{formatDate(val)}
						</span>
					) : (
						<span className="text-sub">—</span>
					),
				),
			},
			{
				id: "title",
				accessorKey: "title",
				header: sortableHeader<JournalActivity>("Title"),
				enableSorting: true,
				cell: valueCell<JournalActivity, string>((val, row) => (
					<Link
						to="/journal/$entryId"
						params={{ entryId: String(row.id) }}
						className="text-[13px] text-ink hover:text-accent-text hover:underline"
					>
						{val}
					</Link>
				)),
			},
			{
				id: "people",
				header: "People",
				cell: ({ row }) => {
					const people = row.original.people ?? [];
					if (!people.length)
						return <span className="text-sub text-[12px]">—</span>;
					return (
						<div className="flex flex-wrap gap-1.5">
							{people.slice(0, 3).map((p) => (
								<PersonChip key={p.person_id} p={p} />
							))}
							{people.length > 3 && (
								<span className="text-[10px] text-sub self-center">
									+{people.length - 3}
								</span>
							)}
						</div>
					);
				},
			},
			{
				id: "content_preview",
				header: "Preview",
				cell: ({ row }) => {
					const content = row.original.content ?? "";
					const preview =
						content.length > 80 ? `${content.slice(0, 80)}…` : content;
					return (
						<span className="text-[12px] text-sub line-clamp-2">
							{preview || "—"}
						</span>
					);
				},
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
				<span className="text-sm text-foreground/50">
					No journal entries yet.
				</span>
			}
		/>
	);
}
