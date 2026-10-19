import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { useTranslation } from "react-i18next";
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
	const { t } = useTranslation("people");
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
		select: (types) => types.filter((rt) => rt.inverse_type_id == null),
	});

	const mutation = useMutation({
		mutationFn: () => connectAllMembers(label.id, Number(typeId)),
		onSuccess: (data) => {
			setResultMsg(t("connectMembers.created", { count: data.created }));
			qc.invalidateQueries({ queryKey: keys.relationships.graph() });
		},
		onError: () => setResultMsg(t("connectMembers.failed")),
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
					<DialogTitle>
						{t("connectMembers.title", { label: label.name })}
					</DialogTitle>
					<DialogDescription>
						{t("connectMembers.description")}
					</DialogDescription>
				</DialogHeader>

				<div className="space-y-4">
					<div>
						<p className="text-sm font-medium text-sub mb-1.5">
							{t("connectMembers.relationshipType")}
						</p>
						{typesQuery.data?.length === 0 ? (
							<p className="text-sm text-sub">
								{t("connectMembers.noSymmetricTypes")}
							</p>
						) : (
							<select
								value={typeId}
								onChange={(e) => setTypeId(e.target.value)}
								className="w-full h-10 border-field-bw border-field-line rounded-md bg-field px-2 text-sm"
							>
								<option value="">{t("connectMembers.select")}</option>
								{typesQuery.data?.map((rt) => (
									<option key={rt.id} value={rt.id}>
										{rt.name}
									</option>
								))}
							</select>
						)}
					</div>

					{previewQuery.isLoading && (
						<p className="text-sm text-sub">
							{t("connectMembers.loadingPreview")}
						</p>
					)}

					{preview && preview.member_count < 2 && (
						<p className="text-sm text-sub">
							{t("connectMembers.tooFew", { members: preview.member_count })}
						</p>
					)}

					{preview && preview.member_count >= 2 && (
						<p className="text-sm text-sub">
							{t("connectMembers.willCreate", {
								pairs: preview.pair_count,
								members: preview.member_count,
							})}
						</p>
					)}

					{resultMsg && (
						<p className="text-sm text-ink font-medium">{resultMsg}</p>
					)}
				</div>

				<DialogFooter>
					<Button variant="neutral" onClick={handleClose}>
						{t("connectMembers.cancel")}
					</Button>
					<Button
						onClick={() => mutation.mutate()}
						disabled={!canConnect || mutation.isPending}
					>
						{mutation.isPending
							? t("connectMembers.connecting")
							: t("connectMembers.connectPairs", {
									pairs: preview?.pair_count ?? "",
								})}
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
