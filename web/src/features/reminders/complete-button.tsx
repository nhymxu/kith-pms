// Complete button: marks a reminder as done and invalidates query cache
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { completeReminder } from "#/endpoints/reminders"
import { keys } from "#/query-keys"
import { Button } from "#/components/ui/button"

interface CompleteButtonProps {
	reminderId: number
	onCompleted?: () => void
}

export function CompleteButton({ reminderId, onCompleted }: CompleteButtonProps) {
	const qc = useQueryClient()

	const mutation = useMutation({
		mutationFn: () => completeReminder(reminderId),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.reminders.all })
			qc.invalidateQueries({ queryKey: keys.reminders.detail(reminderId) })
			onCompleted?.()
		},
	})

	return (
		<Button
			size="sm"
			variant="default"
			onClick={() => mutation.mutate()}
			disabled={mutation.isPending}
		>
			{mutation.isPending ? "…" : "Mark complete"}
		</Button>
	)
}
