// Reminders table: columns for title, person, due date, status
import { useMemo } from "react"
import type { ColumnDef } from "@tanstack/react-table"
import { Link } from "@tanstack/react-router"
import { DataTable } from "#/components/data-table/data-table"
import { sortableHeader, valueCell } from "#/components/data-table/column-helpers"
import type { ReminderWithPerson } from "#/schemas/reminder"
import { CompleteButton } from "./complete-button"

interface RemindersTableProps {
	data: ReminderWithPerson[]
	toolbarActions?: React.ReactNode
	onCompleted?: () => void
}

function StatusBadge({ completed, dueDate }: { completed: boolean; dueDate: string }) {
	if (completed) return <span className="font-mono text-[10px] uppercase text-zinc-400 line-through">Done</span>
	const isOverdue = dueDate ? new Date(dueDate) < new Date() : false
	return isOverdue
		? <span className="font-mono text-[10px] uppercase text-red-600">Overdue</span>
		: <span className="font-mono text-[10px] uppercase text-indigo-600">Upcoming</span>
}

function formatDate(dateStr: string) {
	if (!dateStr) return "—"
	try {
		return new Date(dateStr).toLocaleDateString()
	} catch {
		return dateStr
	}
}

export function RemindersTable({ data, toolbarActions, onCompleted }: RemindersTableProps) {
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
						className="text-[13px] text-zinc-900 hover:text-indigo-600 hover:underline"
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
					val ? <span className="font-mono text-[12px] text-zinc-500">{formatDate(val)}</span> : <span className="text-zinc-300">—</span>
				),
			},
			{
				id: "status",
				header: "Status",
				cell: ({ row }) => (
					<StatusBadge completed={row.original.completed ?? false} dueDate={row.original.due_date} />
				),
			},
			{
				id: "actions",
				header: "",
				cell: ({ row }) =>
					!row.original.completed ? (
						<CompleteButton reminderId={row.original.id} onCompleted={onCompleted} />
					) : null,
			},
		],
		[onCompleted],
	)

	return (
		<DataTable
			columns={columns}
			data={data}
			toolbarActions={toolbarActions}
			emptyState={<span className="text-sm text-foreground/50">No reminders found.</span>}
		/>
	)
}
