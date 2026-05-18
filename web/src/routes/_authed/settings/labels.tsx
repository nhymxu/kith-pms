import { createFileRoute } from "@tanstack/react-router"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { useState } from "react"
import { useForm } from "@tanstack/react-form"
import { listLabels, createLabel, updateLabel, deleteLabel } from "#/endpoints/labels"
import { labelRequestSchema, type Label, type LabelRequest } from "#/schemas/label"
import { keys } from "#/query-keys"
import { Button } from "#/components/ui/button"
import { Badge } from "#/components/ui/badge"
import { FormField } from "#/components/form/form-field"
import { SubmitButton } from "#/components/form/submit-button"
import { Alert, AlertDescription } from "#/components/ui/alert"
import {
	Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter,
} from "#/components/ui/dialog"
import { Pencil, Trash2, Plus } from "lucide-react"

export const Route = createFileRoute("/_authed/settings/labels")({
	component: LabelsPage,
})

// ── Label form (shared by create and edit dialogs) ─────────────────────────

interface LabelFormDialogProps {
	initial?: Label
	onClose: () => void
}

function LabelFormDialog({ initial, onClose }: LabelFormDialogProps) {
	const qc = useQueryClient()
	const [apiError, setApiError] = useState<string | null>(null)

	const mutation = useMutation({
		mutationFn: (body: LabelRequest) =>
			initial ? updateLabel(initial.id, body) : createLabel(body).then(() => undefined),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.labels.all })
			onClose()
		},
		onError: (e) => setApiError(e instanceof Error ? e.message : "Save failed"),
	})

	const form = useForm({
		defaultValues: {
			name: initial?.name ?? "",
			color: initial?.color ?? "#a0c4ff",
		} satisfies LabelRequest,
		validators: {
			onSubmit: ({ value }) => {
				const r = labelRequestSchema.safeParse(value)
				return r.success ? undefined : r.error.issues.map((i) => i.message).join(", ")
			},
		},
		onSubmit: async ({ value }) => mutation.mutateAsync(value as LabelRequest),
	})

	return (
		<form onSubmit={(e) => { e.preventDefault(); form.handleSubmit() }} className="space-y-4">
			{apiError && <Alert variant="destructive"><AlertDescription>{apiError}</AlertDescription></Alert>}
			<form.Field name="name">
				{(f) => <FormField field={f} label="Name *" placeholder="e.g. Family" />}
			</form.Field>
			<form.Field name="color">
				{(f) => <FormField field={f} label="Color (hex)" placeholder="#a0c4ff" type="color" />}
			</form.Field>
			<DialogFooter>
				<Button type="button" variant="neutral" onClick={onClose}>Cancel</Button>
				<form.Subscribe selector={(s) => s.isSubmitting}>
					{(isSubmitting) => (
						<SubmitButton isPending={isSubmitting} pendingLabel="Saving…">
							{initial ? "Save changes" : "Create label"}
						</SubmitButton>
					)}
				</form.Subscribe>
			</DialogFooter>
		</form>
	)
}

// ── Delete confirm dialog ──────────────────────────────────────────────────

function DeleteLabelDialog({ label, onClose }: { label: Label; onClose: () => void }) {
	const qc = useQueryClient()
	const mutation = useMutation({
		mutationFn: () => deleteLabel(label.id),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.labels.all })
			onClose()
		},
	})
	return (
		<>
			<DialogHeader>
				<DialogTitle>Delete label?</DialogTitle>
				<DialogDescription>
					Permanently delete "{label.name}"? People with this label will be unaffected (label detached).
				</DialogDescription>
			</DialogHeader>
			<DialogFooter>
				<Button variant="neutral" onClick={onClose}>Cancel</Button>
				<Button variant="destructive" onClick={() => mutation.mutate()} disabled={mutation.isPending}>
					{mutation.isPending ? "Deleting…" : "Delete"}
				</Button>
			</DialogFooter>
		</>
	)
}

// ── Page ───────────────────────────────────────────────────────────────────

type DialogMode = { kind: "create" } | { kind: "edit"; label: Label } | { kind: "delete"; label: Label } | null

function LabelsPage() {
	const [dialog, setDialog] = useState<DialogMode>(null)
	const { data, isPending } = useQuery({ queryKey: keys.labels.list(), queryFn: listLabels })

	return (
		<div className="space-y-4 max-w-xl">
			<div className="flex items-center justify-between">
				<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">Labels</h1>
				<Button size="sm" onClick={() => setDialog({ kind: "create" })}>
					<Plus className="size-3 mr-1" /> New Label
				</Button>
			</div>

			{isPending && <p className="text-[13px] text-zinc-500">Loading…</p>}

			{data && data.length === 0 && (
				<p className="text-[13px] text-zinc-500">No labels yet. Create one to start categorising people.</p>
			)}

			<ul className="border border-zinc-200 rounded-md bg-white divide-y divide-zinc-100">
				{data?.map((label) => (
					<li key={label.id} className="flex items-center gap-3 px-4 py-3">
						<span className="size-3 rounded-full shrink-0" style={{ backgroundColor: label.color }} />
						<span className="text-[13px] text-zinc-900">{label.name}</span>
						{label.count !== undefined && label.count > 0 && (
							<span className="font-mono text-[11px] text-zinc-400">{label.count} people</span>
						)}
						<div className="ml-auto flex gap-1">
							<Button variant="ghost" size="icon" onClick={() => setDialog({ kind: "edit", label })}>
								<Pencil className="size-3.5" />
							</Button>
							<Button variant="ghost" size="icon" onClick={() => setDialog({ kind: "delete", label })}>
								<Trash2 className="size-3.5" />
							</Button>
						</div>
					</li>
				))}
			</ul>

			{/* Create / Edit dialog */}
			<Dialog
				open={dialog?.kind === "create" || dialog?.kind === "edit"}
				onOpenChange={(v) => !v && setDialog(null)}
			>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>{dialog?.kind === "edit" ? "Edit label" : "New label"}</DialogTitle>
					</DialogHeader>
					<LabelFormDialog
						initial={dialog?.kind === "edit" ? dialog.label : undefined}
						onClose={() => setDialog(null)}
					/>
				</DialogContent>
			</Dialog>

			{/* Delete dialog */}
			<Dialog open={dialog?.kind === "delete"} onOpenChange={(v) => !v && setDialog(null)}>
				<DialogContent>
					{dialog?.kind === "delete" && (
						<DeleteLabelDialog label={dialog.label} onClose={() => setDialog(null)} />
					)}
				</DialogContent>
			</Dialog>
		</div>
	)
}
