import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { getReminder, updateReminder } from "#/endpoints/reminders"
import { keys } from "#/query-keys"
import { ReminderForm } from "#/features/reminders/reminder-form"
import type { ReminderRequest } from "#/schemas/reminder"

export const Route = createFileRoute("/_authed/reminders/$reminderId/edit")({
	component: EditReminderPage,
})

function EditReminderPage() {
	const { reminderId } = Route.useParams()
	const id = Number(reminderId)
	const navigate = useNavigate()
	const qc = useQueryClient()

	const { data, isPending, isError } = useQuery({
		queryKey: keys.reminders.detail(id),
		queryFn: () => getReminder(id),
	})

	const mutation = useMutation({
		mutationFn: (body: ReminderRequest) => updateReminder(id, body),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.reminders.detail(id) })
			qc.invalidateQueries({ queryKey: keys.reminders.all })
			navigate({ to: "/reminders/$reminderId", params: { reminderId } })
		},
	})

	if (isPending) return <p className="text-sm font-base text-foreground/60">Loading…</p>
	if (isError || !data) return <p className="text-sm font-base text-destructive">Reminder not found.</p>

	return (
		<div className="max-w-lg space-y-4">
			<h1 className="text-2xl font-heading">Edit Reminder</h1>
			<ReminderForm
				initial={data}
				onSubmit={(v) => mutation.mutateAsync(v)}
				submitLabel="Update Reminder"
			/>
		</div>
	)
}
