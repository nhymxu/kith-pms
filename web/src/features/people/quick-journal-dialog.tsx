import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
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
import { Textarea } from "#/components/ui/textarea";
import { apiFetch } from "#/lib/api-client";
import { keys } from "#/query-keys";
import { PersonSinglePicker } from "./person-single-picker";

interface QuickJournalDialogProps {
	/** Omit to show a person picker (e.g. dashboard "+ Log interaction"). */
	personId?: number;
	open: boolean;
	onClose: () => void;
}

export function QuickJournalDialog({
	personId,
	open,
	onClose,
}: QuickJournalDialogProps) {
	const [title, setTitle] = useState("");
	const [content, setContent] = useState("");
	const [error, setError] = useState<string | null>(null);
	const [pickedPerson, setPickedPerson] = useState<{
		id: number;
		name: string;
	} | null>(null);
	const qc = useQueryClient();

	const targetPersonId = personId ?? pickedPerson?.id;

	function resetAndClose() {
		setTitle("");
		setContent("");
		setPickedPerson(null);
		onClose();
	}

	const mutation = useMutation({
		mutationFn: () => {
			if (!targetPersonId) {
				throw new Error("Select a person first");
			}
			return apiFetch(`/v1/people/${targetPersonId}/journal/quick`, {
				method: "POST",
				body: JSON.stringify({ title, content }),
			});
		},
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.journal.all });
			// last_contact_at changed: people lists (dashboard widgets) + detail are all under this prefix
			qc.invalidateQueries({ queryKey: keys.people.all });
			resetAndClose();
		},
		onError: (err) => setError(err instanceof Error ? err.message : "Failed"),
	});

	return (
		<Dialog open={open} onOpenChange={(v) => !v && resetAndClose()}>
			<DialogContent>
				<DialogHeader>
					<DialogTitle>Quick journal entry</DialogTitle>
				</DialogHeader>
				{error && (
					<Alert variant="destructive">
						<AlertDescription>{error}</AlertDescription>
					</Alert>
				)}
				<div className="space-y-3">
					{personId === undefined && (
						<div>
							<Label>Person</Label>
							<PersonSinglePicker
								value={pickedPerson}
								onChange={setPickedPerson}
							/>
						</div>
					)}
					<div>
						<Label>Title</Label>
						<Input
							value={title}
							onChange={(e) => setTitle(e.target.value)}
							placeholder="What happened?"
						/>
					</div>
					<div>
						<Label>Notes</Label>
						<Textarea
							value={content}
							onChange={(e) => setContent(e.target.value)}
							rows={3}
							placeholder="Details…"
						/>
					</div>
				</div>
				<DialogFooter>
					<Button variant="neutral" onClick={resetAndClose}>
						Cancel
					</Button>
					<SubmitButton
						isPending={mutation.isPending}
						onClick={() => mutation.mutate()}
						disabled={!targetPersonId}
						type="button"
					>
						Save
					</SubmitButton>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
