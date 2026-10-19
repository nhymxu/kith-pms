import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { SafeTrans } from "#/components/safe-trans";
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
	const { t } = useTranslation("people");
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

	const selectedType = types?.find((rt) => rt.id === Number(typeId));

	const mutation = useMutation({
		mutationFn: () => {
			if (!target) throw new Error("No target person selected");
			return bulkCreateRelationships(
				target.id,
				toIds.map((id) => ({
					to_person_id: id,
					relationship_type_id: Number(typeId),
				})),
			);
		},
		onSuccess: (data) => {
			setResultMsg(
				t("connectTo.connected", {
					created: data.created,
					skipped: data.skipped,
				}),
			);
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
					? t("connectTo.partialFailure", { created: partial.created })
					: t("connectTo.failure"),
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
					<DialogTitle>{t("connectTo.title")}</DialogTitle>
				</DialogHeader>

				<div className="space-y-4">
					<div>
						<p className="text-sm font-medium text-sub mb-1.5">
							{t("connectTo.targetPerson")}
						</p>
						<PersonSinglePicker value={target} onChange={setTarget} />
						{target && toIds.length === 0 && (
							<p className="text-xs text-sub mt-1">{t("connectTo.noOthers")}</p>
						)}
					</div>

					<div>
						<p className="text-sm font-medium text-sub mb-1.5">
							{t("connectTo.relationshipType")}
						</p>
						<select
							value={typeId}
							onChange={(e) => setTypeId(e.target.value)}
							className="w-full h-10 border-field-bw border-field-line rounded-md bg-field px-2 text-sm"
						>
							<option value="">{t("connectTo.selectType")}</option>
							{types?.map((rt) => (
								<option key={rt.id} value={rt.id}>
									{rt.name}
								</option>
							))}
						</select>
						{target && selectedType && (
							<p className="text-xs text-sub mt-1">
								<SafeTrans
									t={t}
									i18nKey="connectTo.willBeSet"
									count={toIds.length}
									values={{
										name: target.name,
										type: selectedType.reverse_name || selectedType.name,
									}}
									components={{ b: <span className="font-medium" /> }}
								/>
							</p>
						)}
					</div>

					{resultMsg && (
						<p className="text-sm text-ink font-medium">{resultMsg}</p>
					)}
				</div>

				<DialogFooter>
					<Button variant="neutral" onClick={handleClose}>
						{t("connectTo.cancel")}
					</Button>
					<Button
						disabled={!canConnect || mutation.isPending}
						onClick={() => mutation.mutate()}
					>
						{mutation.isPending
							? t("connectTo.connecting")
							: t("connectTo.connect", { count: toIds.length })}
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
