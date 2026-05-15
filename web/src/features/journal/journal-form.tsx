// Journal create/edit form — TanStack Form + Zod, people multi-select
import { useForm } from "@tanstack/react-form"
import { useQuery } from "@tanstack/react-query"
import { useState } from "react"
import { journalRequestSchema, type JournalRequest, type JournalActivity } from "#/schemas/journal"
import { listPeople } from "#/endpoints/people"
import { keys } from "#/query-keys"
import { FormField } from "#/components/form/form-field"
import { SubmitButton } from "#/components/form/submit-button"
import { Label } from "#/components/ui/label"
import { Textarea } from "#/components/ui/textarea"
import { Badge } from "#/components/ui/badge"
import { Alert, AlertDescription } from "#/components/ui/alert"
import { X, Plus } from "lucide-react"

interface JournalFormProps {
	initial?: Partial<JournalActivity>
	onSubmit: (values: JournalRequest) => Promise<void>
	submitLabel?: string
	defaultPersonId?: number
}

export function JournalForm({ initial, onSubmit, submitLabel = "Save Entry", defaultPersonId }: JournalFormProps) {
	const [apiError, setApiError] = useState<string | null>(null)
	const [searchQ, setSearchQ] = useState("")

	const { data: peopleList } = useQuery({
		queryKey: keys.people.list({ q: searchQ || undefined }),
		queryFn: () => listPeople({ q: searchQ || undefined, page_size: 50 }),
	})

	const form = useForm({
		defaultValues: {
			title: initial?.title ?? "",
			content: initial?.content ?? "",
			occurred_at_date: initial?.occurred_at_date ?? new Date().toISOString().slice(0, 10),
			occurred_at_time: initial?.occurred_at_time ?? "",
			person_ids: initial?.people?.map((p) => p.person_id) ?? (defaultPersonId ? [defaultPersonId] : []),
		} satisfies JournalRequest,
		validators: {
			onSubmit: ({ value }) => {
				const result = journalRequestSchema.safeParse(value)
				if (!result.success) return result.error.issues.map((i) => i.message).join(", ")
				return undefined
			},
		},
		onSubmit: async ({ value }) => {
			setApiError(null)
			try {
				await onSubmit(value as JournalRequest)
			} catch (err) {
				setApiError(err instanceof Error ? err.message : "Failed to save entry")
			}
		},
	})

	return (
		<form
			onSubmit={(e) => { e.preventDefault(); form.handleSubmit() }}
			className="space-y-4 max-w-2xl"
		>
			{apiError && (
				<Alert variant="destructive">
					<AlertDescription>{apiError}</AlertDescription>
				</Alert>
			)}

			<form.Field name="title">
				{(f) => <FormField field={f} label="Title *" placeholder="What happened?" />}
			</form.Field>

			<div className="grid grid-cols-2 gap-4">
				<form.Field name="occurred_at_date">
					{(f) => <FormField field={f} label="Date *" type="date" />}
				</form.Field>
				<form.Field name="occurred_at_time">
					{(f) => <FormField field={f} label="Time (optional)" type="time" />}
				</form.Field>
			</div>

			{/* People multi-select */}
			<form.Field name="person_ids">
				{(f) => {
					const selectedIds: number[] = Array.isArray(f.state.value) ? f.state.value : []
					const selectedPeople = peopleList?.items.filter((p) => selectedIds.includes(p.id)) ?? []
					const unselected = peopleList?.items.filter((p) => !selectedIds.includes(p.id)) ?? []

					return (
						<div className="space-y-2">
							<Label>People</Label>
							{selectedPeople.length > 0 && (
								<div className="flex flex-wrap gap-1">
									{selectedPeople.map((p) => (
										<Badge key={p.id} variant="neutral" className="flex items-center gap-1">
											{p.name}
											<button
												type="button"
												onClick={() => f.handleChange(selectedIds.filter((id) => id !== p.id))}
												className="ml-0.5 hover:text-destructive"
											>
												<X className="size-3" />
											</button>
										</Badge>
									))}
								</div>
							)}
							<div className="space-y-1">
								<input
									type="text"
									value={searchQ}
									onChange={(e) => setSearchQ(e.target.value)}
									placeholder="Search people to add…"
									className="h-9 w-full border-2 border-border rounded-base bg-background px-3 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
								/>
								{unselected.length > 0 && (
									<div className="flex flex-wrap gap-1 pt-1">
										{unselected.slice(0, 10).map((p) => (
											<button
												key={p.id}
												type="button"
												onClick={() => f.handleChange([...selectedIds, p.id])}
												className="flex items-center gap-1 text-xs border-2 border-dashed border-border rounded-base px-2 py-1 hover:border-main transition-colors"
											>
												<Plus className="size-3" />{p.name}
											</button>
										))}
									</div>
								)}
							</div>
						</div>
					)
				}}
			</form.Field>

			{/* Content */}
			<div className="space-y-1.5">
				<Label>Content</Label>
				<form.Field name="content">
					{(f) => (
						<Textarea
							value={f.state.value}
							onBlur={f.handleBlur}
							onChange={(e) => f.handleChange(e.target.value)}
							placeholder="Write about the interaction…"
							rows={6}
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
