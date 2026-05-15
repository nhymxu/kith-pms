import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { createReminder } from "#/endpoints/reminders"
import { keys } from "#/query-keys"
import { ReminderForm } from "#/features/reminders/reminder-form"
import type { ReminderRequest } from "#/schemas/reminder"

export const Route = createFileRoute("/_authed/reminders/new")({
	component: NewReminderPage,
})

function NewReminderPage() {
	const navigate = useNavigate()
	const qc = useQueryClient()

	const mutation = useMutation({
		mutationFn: (body: ReminderRequest) => createReminder(body).then(() => undefined as void),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.reminders.all })
			navigate({ to: "/reminders" })
		},
	})

	return (
		<div className="max-w-lg space-y-4">
			<h1 className="text-2xl font-heading">New Reminder</h1>
			<ReminderForm onSubmit={(v) => mutation.mutateAsync(v)} submitLabel="Create Reminder" />
		</div>
	)
}
