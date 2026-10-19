import { useMutation, useQueryClient } from "@tanstack/react-query";
import { BookOpen, Clock, Gift } from "lucide-react";
import { useState } from "react";
import { Trans, useTranslation } from "react-i18next";
import { SubmitButton } from "#/components/form/submit-button";
import { Button } from "#/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import { apiFetch } from "#/lib/api-client";
import { formatDateTime } from "#/lib/format-datetime";
import { keys } from "#/query-keys";
import { QuickGiftDialog } from "./quick-gift-dialog";
import { QuickJournalDialog } from "./quick-journal-dialog";

interface QuickActionsProps {
	personId: number;
}

export function QuickActions({ personId }: QuickActionsProps) {
	const { t } = useTranslation("people");
	const [journalOpen, setJournalOpen] = useState(false);
	const [giftOpen, setGiftOpen] = useState(false);
	const [confirmLastContact, setConfirmLastContact] = useState(false);
	const qc = useQueryClient();

	const lastContactMutation = useMutation({
		mutationFn: () =>
			apiFetch(`/v1/people/${personId}/last-contact`, { method: "POST" }),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.people.detail(personId) });
			setConfirmLastContact(false);
		},
	});

	return (
		<>
			<div className="flex flex-wrap gap-2">
				<Button
					variant="neutral"
					size="sm"
					onClick={() => setJournalOpen(true)}
				>
					<BookOpen className="size-3" /> {t("quickActions.journal")}
				</Button>
				<Button variant="neutral" size="sm" onClick={() => setGiftOpen(true)}>
					<Gift className="size-3" /> {t("quickActions.gift")}
				</Button>
				<Button
					variant="neutral"
					size="sm"
					onClick={() => setConfirmLastContact(true)}
				>
					<Clock className="size-3" /> {t("quickActions.updateLastContact")}
				</Button>
			</div>

			<QuickJournalDialog
				personId={personId}
				open={journalOpen}
				onClose={() => setJournalOpen(false)}
			/>
			<QuickGiftDialog
				personId={personId}
				open={giftOpen}
				onClose={() => setGiftOpen(false)}
			/>

			<Dialog
				open={confirmLastContact}
				onOpenChange={(v) => !v && setConfirmLastContact(false)}
			>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>{t("quickActions.confirmTitle")}</DialogTitle>
					</DialogHeader>
					<p className="text-sm text-sub">
						<Trans
							t={t}
							i18nKey="quickActions.confirmBody"
							values={{ date: formatDateTime(new Date().toISOString()) }}
							components={{ b: <span className="font-medium" /> }}
						/>
					</p>
					<DialogFooter>
						<Button
							variant="neutral"
							onClick={() => setConfirmLastContact(false)}
						>
							{t("quickActions.cancel")}
						</Button>
						<SubmitButton
							isPending={lastContactMutation.isPending}
							onClick={() => lastContactMutation.mutate()}
							type="button"
						>
							{t("quickActions.confirm")}
						</SubmitButton>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</>
	);
}
