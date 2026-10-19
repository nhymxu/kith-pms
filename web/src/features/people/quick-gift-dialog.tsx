import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { SubmitButton } from "#/components/form/submit-button";
import { Alert, AlertDescription } from "#/components/ui/alert";
import { Button } from "#/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import { Input } from "#/components/ui/input";
import { Label } from "#/components/ui/label";
import { apiFetch } from "#/lib/api-client";
import { apiErrorMessage } from "#/lib/api-error-message";
import { keys } from "#/query-keys";

export function QuickGiftDialog({
	personId,
	open,
	onClose,
}: {
	personId: number;
	open: boolean;
	onClose: () => void;
}) {
	const { t } = useTranslation("people");
	const [title, setTitle] = useState("");
	const [error, setError] = useState<string | null>(null);
	const qc = useQueryClient();

	const mutation = useMutation({
		mutationFn: () =>
			apiFetch(`/v1/people/${personId}/gifts/quick`, {
				method: "POST",
				body: JSON.stringify({ title }),
			}),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.gifts.all });
			setTitle("");
			onClose();
		},
		onError: (err) => setError(apiErrorMessage(err, t("quickGift.failed"))),
	});

	return (
		<Dialog open={open} onOpenChange={(v) => !v && onClose()}>
			<DialogContent>
				<DialogHeader>
					<DialogTitle>{t("quickGift.title")}</DialogTitle>
				</DialogHeader>
				{error && (
					<Alert variant="destructive">
						<AlertDescription>{error}</AlertDescription>
					</Alert>
				)}
				<div>
					<Label>{t("quickGift.giftTitle")}</Label>
					<Input
						value={title}
						onChange={(e) => setTitle(e.target.value)}
						placeholder={t("quickGift.placeholder")}
					/>
				</div>
				<DialogFooter>
					<Button variant="neutral" onClick={onClose}>
						{t("quickGift.cancel")}
					</Button>
					<SubmitButton
						isPending={mutation.isPending}
						onClick={() => mutation.mutate()}
						type="button"
					>
						{t("quickGift.save")}
					</SubmitButton>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
