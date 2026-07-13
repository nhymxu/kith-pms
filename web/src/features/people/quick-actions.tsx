import { useMutation, useQueryClient } from "@tanstack/react-query";
import { BookOpen, Clock, Gift } from "lucide-react";
import { useState } from "react";
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
					<BookOpen className="size-3" /> Quick journal
				</Button>
				<Button variant="neutral" size="sm" onClick={() => setGiftOpen(true)}>
					<Gift className="size-3" /> Quick gift
				</Button>
				<Button
					variant="neutral"
					size="sm"
					onClick={() => setConfirmLastContact(true)}
				>
					<Clock className="size-3" /> Update last contact to today
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
						<DialogTitle>Update last contact?</DialogTitle>
					</DialogHeader>
					<p className="text-sm text-sub">
						This will set the last contact date to{" "}
						<span className="font-medium">
							{formatDateTime(new Date().toISOString())}
						</span>
						. Continue?
					</p>
					<DialogFooter>
						<Button
							variant="neutral"
							onClick={() => setConfirmLastContact(false)}
						>
							Cancel
						</Button>
						<SubmitButton
							isPending={lastContactMutation.isPending}
							onClick={() => lastContactMutation.mutate()}
							type="button"
						>
							Confirm
						</SubmitButton>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</>
	);
}
