// Reminders table: columns for title, person, due date, recurrence, status

import { Link } from "@tanstack/react-router";
import type { ColumnDef } from "@tanstack/react-table";
import { useMemo } from "react";
import {
	sortableHeader,
	valueCell,
} from "#/components/data-table/column-helpers";
import { DataTable } from "#/components/data-table/data-table";
import { Pill } from "#/components/ui/pill";
import { formatDate, formatDateTime } from "#/lib/format-datetime";
import type { ReminderWithPerson } from "#/schemas/reminder";
import { CompleteButton } from "./complete-button";
import { recurrenceLabel } from "./reminder-form";

// Returns true if the ISO string contains a non-midnight time component.
function hasTime(val: string): boolean {
	if (val.length <= 10) return false;
	const t = val.slice(11, 16);
	return t !== "00:00";
}

interface RemindersTableProps {
	data: ReminderWithPerson[];
	toolbarActions?: React.ReactNode;
	onCompleted?: () => void;
}

function StatusBadge({
	completed,
	dueDate,
}: {
	completed: boolean;
	dueDate: string;
}) {
	if (completed)
		return (
			<Pill variant="plain" strike>
				Done
			</Pill>
		);
	const isOverdue = dueDate ? new Date(dueDate) < new Date() : false;
	return isOverdue ? (
		<Pill variant="danger">Overdue</Pill>
	) : (
		<Pill variant="accent">Upcoming</Pill>
	);
}

export function RemindersTable({
	data,
	toolbarActions,
	onCompleted,
}: RemindersTableProps) {
	const columns = useMemo<ColumnDef<ReminderWithPerson>[]>(
		() => [
			{
				id: "title",
				accessorKey: "title",
				header: sortableHeader<ReminderWithPerson>("Title"),
				enableSorting: true,
				cell: valueCell<ReminderWithPerson, string>((val, row) => (
					<Link
						to="/reminders/$reminderId"
						params={{ reminderId: String(row.id) }}
						className="text-[13px] text-ink hover:text-accent-text hover:underline"
					>
						{val}
					</Link>
				)),
			},
			{
				id: "person_name",
				accessorKey: "person_name",
				header: sortableHeader<ReminderWithPerson>("Person"),
				enableSorting: true,
				cell: valueCell<ReminderWithPerson, string>((val) => val || "—"),
			},
			{
				id: "due_date",
				accessorKey: "due_date",
				header: sortableHeader<ReminderWithPerson>("Due Date"),
				enableSorting: true,
				cell: valueCell<ReminderWithPerson, string>((val) =>
					val ? (
						<span className="font-mono text-[12px] text-sub">
							{hasTime(val) ? formatDateTime(val) : formatDate(val)}
						</span>
					) : (
						<span className="text-sub/60">—</span>
					),
				),
			},
			{
				id: "recurrence",
				header: "Recurrence",
				cell: ({ row }) =>
					row.original.recurrence_rule ? (
						<Pill variant="accent">
							↻ {recurrenceLabel(row.original.recurrence_rule)}
						</Pill>
					) : (
						<span className="text-sub/60">—</span>
					),
			},
			{
				id: "status",
				header: "Status",
				cell: ({ row }) => (
					<StatusBadge
						completed={row.original.completed ?? false}
						dueDate={row.original.due_date}
					/>
				),
			},
			{
				id: "actions",
				header: "",
				cell: ({ row }) =>
					!row.original.completed ? (
						<CompleteButton
							reminderId={row.original.id}
							onCompleted={onCompleted}
						/>
					) : null,
			},
		],
		[onCompleted],
	);

	return (
		<DataTable
			columns={columns}
			data={data}
			toolbarActions={toolbarActions}
			emptyState={
				<span className="text-sm text-foreground/50">No reminders found.</span>
			}
		/>
	);
}
