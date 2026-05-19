// Gift create/edit form — TanStack Form + Zod validation
import { useForm } from "@tanstack/react-form"
import { useQuery } from "@tanstack/react-query"
import { giftRequestSchema, type GiftRequest, type Gift } from "#/schemas/gift"
import { listPeople } from "#/endpoints/people"
import { keys } from "#/query-keys"
import { FormField } from "#/components/form/form-field"
import { SubmitButton } from "#/components/form/submit-button"
import { Label } from "#/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "#/components/ui/select"
import { Textarea } from "#/components/ui/textarea"
import { Alert, AlertDescription } from "#/components/ui/alert"
import { useState } from "react"

interface GiftFormProps {
	initial?: Partial<Gift>
	onSubmit: (values: GiftRequest) => Promise<void>
	submitLabel?: string
}

export function GiftForm({ initial, onSubmit, submitLabel = "Save Gift" }: GiftFormProps) {
	const [apiError, setApiError] = useState<string | null>(null)

	const { data: peopleList } = useQuery({
		queryKey: keys.people.list({}),
		queryFn: () => listPeople({ page_size: 200 }),
	})

	const form = useForm({
		defaultValues: {
			person_id: initial?.person_id ?? 0,
			title: initial?.title ?? "",
			direction: initial?.direction ?? "planned",
			date: initial?.date ?? "",
			notes: initial?.notes ?? "",
			amount_cents: initial?.amount_cents ?? null,
			currency: initial?.currency ?? "USD",
			debt_type: initial?.debt_type ?? "",
		} satisfies GiftRequest,
		validators: {
			onSubmit: ({ value }) => {
				const result = giftRequestSchema.safeParse(value)
				if (!result.success) {
					return result.error.issues.map((i) => i.message).join(", ")
				}
				return undefined
			},
		},
		onSubmit: async ({ value }) => {
			setApiError(null)
			try {
				await onSubmit(value as GiftRequest)
			} catch (err) {
				setApiError(err instanceof Error ? err.message : "Failed to save gift")
			}
		},
	})

	return (
		<form
			onSubmit={(e) => {
				e.preventDefault()
				form.handleSubmit()
			}}
			className="space-y-4"
		>
			{apiError && (
				<Alert variant="destructive">
					<AlertDescription>{apiError}</AlertDescription>
				</Alert>
			)}

			{/* Person selector */}
			<div className="space-y-1.5">
				<Label>Person</Label>
				<form.Field name="person_id">
					{(field) => (
						<Select
							value={field.state.value ? String(field.state.value) : ""}
							onValueChange={(v) => field.handleChange(Number(v))}
						>
							<SelectTrigger>
								<SelectValue placeholder="Select person…" />
							</SelectTrigger>
							<SelectContent>
								{peopleList?.items.map((p) => (
									<SelectItem key={p.id} value={String(p.id)}>
										{p.name}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					)}
				</form.Field>
			</div>

			<form.Field name="title">
				{(field) => <FormField field={field} label="Title / Occasion" placeholder="Birthday gift" />}
			</form.Field>

			{/* Direction */}
			<div className="space-y-1.5">
				<Label>Direction</Label>
				<form.Field name="direction">
					{(field) => (
						<Select value={field.state.value} onValueChange={(v) => field.handleChange(v as GiftRequest["direction"])}>
							<SelectTrigger>
								<SelectValue placeholder="Select direction…" />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="given">Given</SelectItem>
								<SelectItem value="received">Received</SelectItem>
								<SelectItem value="planned">Planned</SelectItem>
							</SelectContent>
						</Select>
					)}
				</form.Field>
			</div>

			{/* Debt type */}
			<div className="space-y-1.5">
				<Label>Debt type</Label>
				<form.Field name="debt_type">
					{(field) => (
						<Select
								value={field.state.value === "" || field.state.value == null ? "none" : field.state.value}
								onValueChange={(v) => field.handleChange((v === "none" ? "" : v) as GiftRequest["debt_type"])}
							>
							<SelectTrigger>
								<SelectValue placeholder="None" />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="none">None</SelectItem>
								<SelectItem value="i_owe">I owe</SelectItem>
								<SelectItem value="they_owe">They owe</SelectItem>
							</SelectContent>
						</Select>
					)}
				</form.Field>
			</div>

			<form.Field name="date">
				{(field) => (
					<FormField field={field} label="Date" type="date" />
				)}
			</form.Field>

			{/* Notes */}
			<div className="space-y-1.5">
				<Label>Notes</Label>
				<form.Field name="notes">
					{(field) => (
						<Textarea
							value={field.state.value}
							onBlur={field.handleBlur}
							onChange={(e) => field.handleChange(e.target.value)}
							placeholder="Optional notes…"
							rows={3}
						/>
					)}
				</form.Field>
			</div>

			<form.Subscribe selector={(s) => s.isSubmitting}>
				{(isSubmitting) => (
					<SubmitButton isPending={isSubmitting} pendingLabel="Saving…">
						{submitLabel}
					</SubmitButton>
				)}
			</form.Subscribe>
		</form>
	)
}
