import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { Button } from "#/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import { listRelationshipTypes } from "#/endpoints/relationship-types";
import { bulkCreateRelationships } from "#/endpoints/relationships";
import { keys } from "#/query-keys";
import { PersonSinglePicker } from "./person-single-picker";

interface ConnectToPersonDialogProps {
	selectedIds: number[];
	open: boolean;
	onClose: () => void;
	onSuccess: () => void;
}

export function ConnectToPersonDialog({
	selectedIds,
	open,
	onClose,
	onSuccess,
}: ConnectToPersonDialogProps) {
	const qc = useQueryClient();
	const [target, setTarget] = useState<{ id: number; name: string } | null>(
		null,
	);
	const [typeId, setTypeId] = useState<string>("");
	const [resultMsg, setResultMsg] = useState<string | null>(null);

	const { data: types } = useQuery({
		queryKey: keys.relationshipTypes.list(),
		queryFn: listRelationshipTypes,
		enabled: open,
	});

	const toIds = target
		? selectedIds.filter((id) => id !== target.id)
		: selectedIds;

	const selectedType = types?.find((t) => t.id === Number(typeId));

	const mutation = useMutation({
		mutationFn: () =>
			bulkCreateRelationships(
				target!.id,
				toIds.map((id) => ({
					to_person_id: id,
					relationship_type_id: Number(typeId),
				})),
			),
		onSuccess: (data) => {
			setResultMsg(`Connected ${data.created} (${data.skipped} skipped)`);
			qc.invalidateQueries({ queryKey: keys.relationships.graph() });
			qc.invalidateQueries({ queryKey: keys.people.all });
			qc.invalidateQueries({ queryKey: keys.relationshipTypes.list() });
			onSuccess();
		},
		onError: (
			err: Error & { partial?: { created: number; skipped: number } },
		) => {
			const partial = err.partial;
			setResultMsg(
				partial
					? `Failed after connecting ${partial.created}`
					: "Failed to connect",
			);
		},
	});

	function handleClose() {
		setTarget(null);
		setTypeId("");
		setResultMsg(null);
		onClose();
	}

	const canConnect = !!target && !!typeId && toIds.length > 0;

	return (
		<Dialog open={open} onOpenChange={(v) => !v && handleClose()}>
			<DialogContent className="max-w-md">
				<DialogHeader>
					<DialogTitle>Connect to person</DialogTitle>
				</DialogHeader>

				<div className="space-y-4">
					<div>
						<p className="text-sm font-medium text-zinc-700 mb-1.5">
							Target person
						</p>
						<PersonSinglePicker value={target} onChange={setTarget} />
						{target && toIds.length === 0 && (
							<p className="text-xs text-zinc-500 mt-1">
								No other people to connect — target is the only selection.
							</p>
						)}
					</div>

					<div>
						<p className="text-sm font-medium text-zinc-700 mb-1.5">
							Relationship type
						</p>
						<select
							value={typeId}
							onChange={(e) => setTypeId(e.target.value)}
							className="w-full h-10 border border-zinc-200 rounded-md bg-white px-2 text-sm"
						>
							<option value="">Select type…</option>
							{types?.map((t) => (
								<option key={t.id} value={t.id}>
									{t.name}
								</option>
							))}
						</select>
						{target && selectedType && (
							<p className="text-xs text-zinc-500 mt-1">
								{target.name} will be set as{" "}
								<span className="font-medium">
									{selectedType.reverse_name || selectedType.name}
								</span>{" "}
								of the {toIds.length} {toIds.length === 1 ? "person" : "people"}
								.
							</p>
						)}
					</div>

					{resultMsg && (
						<p className="text-sm text-zinc-700 font-medium">{resultMsg}</p>
					)}
				</div>

				<DialogFooter>
					<Button variant="neutral" onClick={handleClose}>
						Cancel
					</Button>
					<Button
						disabled={!canConnect || mutation.isPending}
						onClick={() => mutation.mutate()}
					>
						{mutation.isPending
							? "Connecting…"
							: `Connect ${toIds.length} ${toIds.length === 1 ? "person" : "people"}`}
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
