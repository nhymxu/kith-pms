// Complete button: marks a reminder as done and invalidates query cache
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { Button } from "#/components/ui/button";
import { completeReminder } from "#/endpoints/reminders";
import { keys } from "#/query-keys";

interface CompleteButtonProps {
	reminderId: number;
	onCompleted?: () => void;
}

export function CompleteButton({
	reminderId,
	onCompleted,
}: CompleteButtonProps) {
	const { t } = useTranslation("reminders");
	const qc = useQueryClient();

	const mutation = useMutation({
		mutationFn: () => completeReminder(reminderId),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.reminders.all });
			qc.invalidateQueries({ queryKey: keys.reminders.detail(reminderId) });
			onCompleted?.();
		},
	});

	return (
		<Button
			size="sm"
			variant="default"
			onClick={() => mutation.mutate()}
			disabled={mutation.isPending}
		>
			{mutation.isPending ? "…" : t("complete.action")}
		</Button>
	);
}
