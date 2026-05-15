// Reminders table: columns for title, person, due date, status
import { useMemo } from "react"
import type { ColumnDef } from "@tanstack/react-table"
import { Link } from "@tanstack/react-router"
import { DataTable } from "#/components/data-table/data-table"
import { Badge } from "#/components/ui/badge"
import { sortableHeader, valueCell } from "#/components/data-table/column-helpers"
import type { ReminderWithPerson } from "#/schemas/reminder"
import { CompleteButton } from "./complete-button"

interface RemindersTableProps {
	data: ReminderWithPerson[]
	toolbarActions?: React.ReactNode
	onCompleted?: () => void
}

function StatusBadge({ completed, dueDate }: { completed: boolean; dueDate: string }) {
	if (completed) return <Badge variant="neutral">Done</Badge>
	const isOverdue = dueDate ? new Date(dueDate) < new Date() : false
	return isOverdue
		? <Badge className="bg-red-300 text-black border-black">Overdue</Badge>
		: <Badge>Upcoming</Badge>
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
						className="font-base underline hover:text-main"
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
				cell: valueCell<ReminderWithPerson, string>((val) => formatDate(val)),
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
