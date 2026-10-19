// Reminders table: columns for title, person, due date, recurrence, status

import { Link } from "@tanstack/react-router";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import {
	sortableHeader,
	type TableColumn,
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
	const { t } = useTranslation("reminders");
	if (completed)
		return (
			<Pill variant="plain" strike>
				{t("status.done")}
			</Pill>
		);
	const isOverdue = dueDate ? new Date(dueDate) < new Date() : false;
	return isOverdue ? (
		<Pill variant="danger">{t("status.overdue")}</Pill>
	) : (
		<Pill variant="accent">{t("status.upcoming")}</Pill>
	);
}

export function RemindersTable({
	data,
	toolbarActions,
	onCompleted,
}: RemindersTableProps) {
	const { t } = useTranslation("reminders");
	const columns = useMemo<TableColumn<ReminderWithPerson>[]>(
		() => [
			{
				id: "title",
				accessorKey: "title",
				header: sortableHeader<ReminderWithPerson>(t("table.title")),
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
				header: sortableHeader<ReminderWithPerson>(t("table.person")),
				enableSorting: true,
				cell: valueCell<ReminderWithPerson, string>((val) => val || "—"),
			},
			{
				id: "due_date",
				accessorKey: "due_date",
				header: sortableHeader<ReminderWithPerson>(t("table.dueDate")),
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
				header: t("table.recurrence"),
				cell: ({ row }) =>
					row.original.recurrence_rule ? (
						<Pill variant="accent">
							↻ {recurrenceLabel(row.original.recurrence_rule, t)}
						</Pill>
					) : (
						<span className="text-sub/60">—</span>
					),
			},
			{
				id: "status",
				header: t("table.status"),
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
		[onCompleted, t],
	);

	return (
		<DataTable
			columns={columns}
			data={data}
			toolbarActions={toolbarActions}
			emptyState={
				<span className="text-sm text-foreground/50">{t("list.empty")}</span>
			}
		/>
	);
}
