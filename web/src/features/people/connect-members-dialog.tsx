import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { Button } from "#/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import {
	connectAllMembers,
	previewConnectAll,
} from "#/endpoints/people-labels";
import { listRelationshipTypes } from "#/endpoints/relationship-types";
import { keys } from "#/query-keys";
import type { PeopleLabel } from "#/schemas/people-label";

interface ConnectMembersDialogProps {
	label: PeopleLabel;
	open: boolean;
	onClose: () => void;
}

export function ConnectMembersDialog({
	label,
	open,
	onClose,
}: ConnectMembersDialogProps) {
	const qc = useQueryClient();
	const [typeId, setTypeId] = useState<string>("");
	const [resultMsg, setResultMsg] = useState<string | null>(null);

	const previewQuery = useQuery({
		queryKey: keys.peopleLabels.connectPreview(label.id),
		queryFn: () => previewConnectAll(label.id),
		enabled: open,
	});

	const typesQuery = useQuery({
		queryKey: keys.relationshipTypes.list(),
		queryFn: listRelationshipTypes,
		enabled: open,
		select: (types) => types.filter((t) => t.inverse_type_id == null),
	});

	const mutation = useMutation({
		mutationFn: () => connectAllMembers(label.id, Number(typeId)),
		onSuccess: (data) => {
			setResultMsg(`Created ${data.created} relationships`);
			qc.invalidateQueries({ queryKey: keys.relationships.graph() });
		},
		onError: () => setResultMsg("Failed to connect members"),
	});

	const preview = previewQuery.data;
	const canConnect =
		typeId !== "" && preview !== undefined && preview.member_count >= 2;

	function handleClose() {
		setTypeId("");
		setResultMsg(null);
		onClose();
	}

	return (
		<Dialog open={open} onOpenChange={(v) => !v && handleClose()}>
			<DialogContent>
				<DialogHeader>
					<DialogTitle>Connect members of "{label.name}"</DialogTitle>
					<DialogDescription>
						Creates relationships between all members of this label.
					</DialogDescription>
				</DialogHeader>

				<div className="space-y-4">
					<div>
						<p className="text-sm font-medium text-zinc-700 mb-1.5">
							Relationship type
						</p>
						{typesQuery.data?.length === 0 ? (
							<p className="text-sm text-zinc-500">
								No symmetric relationship types available. Create one (e.g.
								Co-worker) in Settings → Relationship Types.
							</p>
						) : (
							<select
								value={typeId}
								onChange={(e) => setTypeId(e.target.value)}
								className="w-full h-10 border border-zinc-200 rounded-md bg-white px-2 text-sm"
							>
								<option value="">Select…</option>
								{typesQuery.data?.map((t) => (
									<option key={t.id} value={t.id}>
										{t.name}
									</option>
								))}
							</select>
						)}
					</div>

					{previewQuery.isLoading && (
						<p className="text-sm text-zinc-400">Loading preview…</p>
					)}

					{preview && preview.member_count < 2 && (
						<p className="text-sm text-zinc-500">
							Need at least 2 members to connect. This label has{" "}
							{preview.member_count}.
						</p>
					)}

					{preview && preview.member_count >= 2 && (
						<p className="text-sm text-zinc-500">
							Will create up to {preview.pair_count} relationships between{" "}
							{preview.member_count} people. Existing relationships are skipped.
						</p>
					)}

					{resultMsg && (
						<p className="text-sm text-zinc-700 font-medium">{resultMsg}</p>
					)}
				</div>

				<DialogFooter>
					<Button variant="neutral" onClick={handleClose}>
						Cancel
					</Button>
					<Button
						onClick={() => mutation.mutate()}
						disabled={!canConnect || mutation.isPending}
					>
						{mutation.isPending
							? "Connecting…"
							: `Connect ${preview?.pair_count ?? ""} pairs`}
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
