import { useForm } from "@tanstack/react-form";
import {
	useMutation,
	useQueryClient,
	useSuspenseQuery,
} from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Network, Pencil, Plus, Trash2 } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { FormField } from "#/components/form/form-field";
import { SubmitButton } from "#/components/form/submit-button";
import { QueryBoundary } from "#/components/query-boundary";
import { Alert, AlertDescription } from "#/components/ui/alert";
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
	createPeopleLabel,
	deletePeopleLabel,
	listPeopleLabels,
	updatePeopleLabel,
} from "#/endpoints/people-labels";
import { ConnectMembersDialog } from "#/features/people/connect-members-dialog";
import { apiErrorMessage } from "#/lib/api-error-message";
import { keys } from "#/query-keys";
import {
	type PeopleLabel,
	type PeopleLabelRequest,
	peopleLabelRequestSchema,
} from "#/schemas/people-label";

export const Route = createFileRoute("/_authed/settings/_layout/people-labels")(
	{
		component: PeopleLabelsPage,
	},
);

interface LabelFormDialogProps {
	initial?: PeopleLabel;
	onClose: () => void;
}

function LabelFormDialog({ initial, onClose }: LabelFormDialogProps) {
	const { t } = useTranslation("settings");
	const qc = useQueryClient();
	const [apiError, setApiError] = useState<string | null>(null);

	const mutation = useMutation({
		mutationFn: (body: PeopleLabelRequest) =>
			initial
				? updatePeopleLabel(initial.id, body)
				: createPeopleLabel(body).then(() => undefined),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.peopleLabels.all });
			onClose();
		},
		onError: (e) => setApiError(apiErrorMessage(e, t("common.saveError"))),
	});

	const form = useForm({
		defaultValues: {
			name: initial?.name ?? "",
			color: initial?.color ?? "#a0c4ff",
		} satisfies PeopleLabelRequest,
		validators: {
			onSubmit: ({ value }) => {
				const r = peopleLabelRequestSchema.safeParse(value);
				return r.success
					? undefined
					: r.error.issues.map((i) => i.message).join(", ");
			},
		},
		onSubmit: async ({ value }) =>
			mutation.mutateAsync(value as PeopleLabelRequest),
	});

	return (
		<form
			onSubmit={(e) => {
				e.preventDefault();
				form.handleSubmit();
			}}
			className="space-y-4"
		>
			{apiError && (
				<Alert variant="destructive">
					<AlertDescription>{apiError}</AlertDescription>
				</Alert>
			)}
			<form.Field name="name">
				{(f) => (
					<FormField
						field={f}
						label={t("peopleLabels.form.name")}
						placeholder={t("peopleLabels.form.namePlaceholder")}
					/>
				)}
			</form.Field>
			<form.Field name="color">
				{(f) => (
					<FormField
						field={f}
						label={t("peopleLabels.form.color")}
						placeholder="#a0c4ff"
						type="color"
					/>
				)}
			</form.Field>
			<DialogFooter>
				<Button type="button" variant="neutral" onClick={onClose}>
					{t("common.cancel")}
				</Button>
				<form.Subscribe selector={(s) => s.isSubmitting}>
					{(isSubmitting) => (
						<SubmitButton
							isPending={isSubmitting}
							pendingLabel={t("common.saving")}
						>
							{initial
								? t("common.saveChanges")
								: t("peopleLabels.form.create")}
						</SubmitButton>
					)}
				</form.Subscribe>
			</DialogFooter>
		</form>
	);
}

function DeleteLabelDialog({
	label,
	onClose,
}: {
	label: PeopleLabel;
	onClose: () => void;
}) {
	const { t } = useTranslation("settings");
	const qc = useQueryClient();
	const mutation = useMutation({
		mutationFn: () => deletePeopleLabel(label.id),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.peopleLabels.all });
			onClose();
		},
	});
	return (
		<>
			<DialogHeader>
				<DialogTitle>{t("peopleLabels.delete.title")}</DialogTitle>
				<DialogDescription>
					{t("peopleLabels.delete.description", { name: label.name })}
				</DialogDescription>
			</DialogHeader>
			<DialogFooter>
				<Button variant="neutral" onClick={onClose}>
					{t("common.cancel")}
				</Button>
				<Button
					variant="destructive"
					onClick={() => mutation.mutate()}
					disabled={mutation.isPending}
				>
					{mutation.isPending ? t("common.deleting") : t("common.delete")}
				</Button>
			</DialogFooter>
		</>
	);
}

type DialogMode =
	| { kind: "create" }
	| { kind: "edit"; label: PeopleLabel }
	| { kind: "delete"; label: PeopleLabel }
	| null;

interface PeopleLabelsListProps {
	onDialogChange: (dialog: DialogMode) => void;
	onConnectingLabelChange: (label: PeopleLabel) => void;
}

function PeopleLabelsList({
	onDialogChange,
	onConnectingLabelChange,
}: PeopleLabelsListProps) {
	const { t } = useTranslation("settings");
	const { data } = useSuspenseQuery({
		queryKey: keys.peopleLabels.list(),
		queryFn: listPeopleLabels,
	});

	return (
		<>
			{data.length === 0 && (
				<p className="text-[13px] text-sub">{t("peopleLabels.empty")}</p>
			)}

			<ul className="border-bw border-line rounded-base bg-panel divide-y divide-line-soft">
				{data.map((label) => (
					<li key={label.id} className="flex items-center gap-3 px-4 py-3">
						<span
							className="size-3 rounded-full shrink-0"
							style={{ backgroundColor: label.color }}
						/>
						<span className="text-[13px] text-ink">{label.name}</span>
						{label.count !== undefined && label.count > 0 && (
							<span className="font-mono text-[11px] text-sub">
								{t("peopleLabels.count", { count: label.count })}
							</span>
						)}
						<div className="ml-auto flex gap-1">
							<Button
								variant="ghost"
								size="icon"
								title={t("peopleLabels.connectMembers")}
								onClick={() => onConnectingLabelChange(label)}
							>
								<Network className="size-3.5" />
							</Button>
							<Button
								variant="ghost"
								size="icon"
								onClick={() => onDialogChange({ kind: "edit", label })}
							>
								<Pencil className="size-3.5" />
							</Button>
							<Button
								variant="ghost"
								size="icon"
								onClick={() => onDialogChange({ kind: "delete", label })}
							>
								<Trash2 className="size-3.5" />
							</Button>
						</div>
					</li>
				))}
			</ul>
		</>
	);
}

function PeopleLabelsPage() {
	const { t } = useTranslation("settings");
	const [dialog, setDialog] = useState<DialogMode>(null);
	const [connectingLabel, setConnectingLabel] = useState<PeopleLabel | null>(
		null,
	);

	return (
		<div className="space-y-4 max-w-xl">
			<div className="flex items-center justify-between">
				<h1 className="text-[18px] font-semibold tracking-tight text-ink font-display">
					{t("peopleLabels.title")}
				</h1>
				<Button size="sm" onClick={() => setDialog({ kind: "create" })}>
					<Plus className="size-3 mr-1" /> {t("peopleLabels.new")}
				</Button>
			</div>

			<QueryBoundary>
				<PeopleLabelsList
					onDialogChange={setDialog}
					onConnectingLabelChange={setConnectingLabel}
				/>
			</QueryBoundary>

			<Dialog
				open={dialog?.kind === "create" || dialog?.kind === "edit"}
				onOpenChange={(v) => !v && setDialog(null)}
			>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>
							{dialog?.kind === "edit"
								? t("peopleLabels.dialog.editTitle")
								: t("peopleLabels.dialog.newTitle")}
						</DialogTitle>
					</DialogHeader>
					<LabelFormDialog
						initial={dialog?.kind === "edit" ? dialog.label : undefined}
						onClose={() => setDialog(null)}
					/>
				</DialogContent>
			</Dialog>

			<Dialog
				open={dialog?.kind === "delete"}
				onOpenChange={(v) => !v && setDialog(null)}
			>
				<DialogContent>
					{dialog?.kind === "delete" && (
						<DeleteLabelDialog
							label={dialog.label}
							onClose={() => setDialog(null)}
						/>
					)}
				</DialogContent>
			</Dialog>

			{connectingLabel && (
				<ConnectMembersDialog
					label={connectingLabel}
					open={true}
					onClose={() => setConnectingLabel(null)}
				/>
			)}
		</div>
	);
}
