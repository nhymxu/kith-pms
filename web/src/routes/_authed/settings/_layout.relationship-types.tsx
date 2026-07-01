import { useForm } from "@tanstack/react-form";
import {
	useMutation,
	useQueryClient,
	useSuspenseQuery,
} from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Pencil, Plus, Trash2 } from "lucide-react";
import { useState } from "react";
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
	createRelationshipType,
	deleteRelationshipType,
	listRelationshipTypes,
	updateRelationshipType,
} from "#/endpoints/relationship-types";
import { keys } from "#/query-keys";
import {
	type RelationshipType,
	type RelationshipTypeRequest,
	relationshipTypeRequestSchema,
} from "#/schemas/relationship-type";

export const Route = createFileRoute(
	"/_authed/settings/_layout/relationship-types",
)({
	component: RelationshipTypesPage,
});

interface RelTypeFormDialogProps {
	initial?: RelationshipType;
	onClose: () => void;
}

function RelTypeFormDialog({ initial, onClose }: RelTypeFormDialogProps) {
	const qc = useQueryClient();
	const [apiError, setApiError] = useState<string | null>(null);

	const mutation = useMutation({
		mutationFn: (body: RelationshipTypeRequest) =>
			initial
				? updateRelationshipType(initial.id, body)
				: createRelationshipType(body).then(() => undefined),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.relationshipTypes.all });
			onClose();
		},
		onError: (e) => setApiError(e instanceof Error ? e.message : "Save failed"),
	});

	const form = useForm({
		defaultValues: {
			name: initial?.name ?? "",
			reverse_name: initial?.reverse_name ?? "",
		} satisfies RelationshipTypeRequest,
		validators: {
			onSubmit: ({ value }) => {
				const r = relationshipTypeRequestSchema.safeParse(value);
				return r.success
					? undefined
					: r.error.issues.map((i) => i.message).join(", ");
			},
		},
		onSubmit: async ({ value }) =>
			mutation.mutateAsync(value as RelationshipTypeRequest),
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
					<FormField field={f} label="Name *" placeholder="e.g. Friend" />
				)}
			</form.Field>
			<form.Field name="reverse_name">
				{(f) => (
					<FormField
						field={f}
						label="Reverse name (optional)"
						placeholder="e.g. Friend of"
					/>
				)}
			</form.Field>
			<DialogFooter>
				<Button type="button" variant="neutral" onClick={onClose}>
					Cancel
				</Button>
				<form.Subscribe selector={(s) => s.isSubmitting}>
					{(isSubmitting) => (
						<SubmitButton isPending={isSubmitting} pendingLabel="Saving…">
							{initial ? "Save changes" : "Create type"}
						</SubmitButton>
					)}
				</form.Subscribe>
			</DialogFooter>
		</form>
	);
}

function DeleteRelTypeDialog({
	relType,
	onClose,
}: {
	relType: RelationshipType;
	onClose: () => void;
}) {
	const qc = useQueryClient();
	const mutation = useMutation({
		mutationFn: () => deleteRelationshipType(relType.id),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.relationshipTypes.all });
			onClose();
		},
	});
	return (
		<>
			<DialogHeader>
				<DialogTitle>Delete relationship type?</DialogTitle>
				<DialogDescription>
					Permanently delete "{relType.name}"? Existing relationships of this
					type will be affected.
				</DialogDescription>
			</DialogHeader>
			<DialogFooter>
				<Button variant="neutral" onClick={onClose}>
					Cancel
				</Button>
				<Button
					variant="destructive"
					onClick={() => mutation.mutate()}
					disabled={mutation.isPending}
				>
					{mutation.isPending ? "Deleting…" : "Delete"}
				</Button>
			</DialogFooter>
		</>
	);
}

type DialogMode =
	| { kind: "create" }
	| { kind: "edit"; relType: RelationshipType }
	| { kind: "delete"; relType: RelationshipType }
	| null;

interface RelationshipTypesListProps {
	onDialogChange: (dialog: DialogMode) => void;
}

function RelationshipTypesList({ onDialogChange }: RelationshipTypesListProps) {
	const { data } = useSuspenseQuery({
		queryKey: keys.relationshipTypes.list(),
		queryFn: listRelationshipTypes,
	});

	return (
		<>
			{data.length === 0 && (
				<p className="text-[13px] text-zinc-500">No relationship types yet.</p>
			)}

			<ul className="border border-zinc-200 rounded-md bg-white divide-y divide-zinc-100">
				{data.map((rt) => (
					<li
						key={rt.id}
						className="flex items-center gap-3 px-4 py-3 text-[13px]"
					>
						<span className="text-zinc-900 font-medium">{rt.name}</span>
						{rt.reverse_name && (
							<span className="text-zinc-400 font-mono text-[11px]">
								↔ {rt.reverse_name}
							</span>
						)}
						{(rt.usage_count ?? 0) > 0 && (
							<span className="font-mono text-[11px] text-zinc-400">
								{rt.usage_count} uses
							</span>
						)}
						<div className="ml-auto flex gap-1">
							<Button
								variant="ghost"
								size="icon"
								onClick={() => onDialogChange({ kind: "edit", relType: rt })}
							>
								<Pencil className="size-3.5" />
							</Button>
							<Button
								variant="ghost"
								size="icon"
								onClick={() => onDialogChange({ kind: "delete", relType: rt })}
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

function RelationshipTypesPage() {
	const [dialog, setDialog] = useState<DialogMode>(null);

	return (
		<div className="space-y-4 max-w-xl">
			<div className="flex items-center justify-between">
				<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">
					Relationship Types
				</h1>
				<Button size="sm" onClick={() => setDialog({ kind: "create" })}>
					<Plus className="size-3 mr-1" /> New Type
				</Button>
			</div>

			<QueryBoundary>
				<RelationshipTypesList onDialogChange={setDialog} />
			</QueryBoundary>

			<Dialog
				open={dialog?.kind === "create" || dialog?.kind === "edit"}
				onOpenChange={(v) => !v && setDialog(null)}
			>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>
							{dialog?.kind === "edit"
								? "Edit relationship type"
								: "New relationship type"}
						</DialogTitle>
					</DialogHeader>
					<RelTypeFormDialog
						initial={dialog?.kind === "edit" ? dialog.relType : undefined}
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
						<DeleteRelTypeDialog
							relType={dialog.relType}
							onClose={() => setDialog(null)}
						/>
					)}
				</DialogContent>
			</Dialog>
		</div>
	);
}
