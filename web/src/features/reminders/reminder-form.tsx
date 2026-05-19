// Reminder create/edit form — TanStack Form + Zod validation
import { useForm } from "@tanstack/react-form";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { FormField } from "#/components/form/form-field";
import { SubmitButton } from "#/components/form/submit-button";
import { Alert, AlertDescription } from "#/components/ui/alert";
import { Label } from "#/components/ui/label";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "#/components/ui/select";
import { Textarea } from "#/components/ui/textarea";
import { listPeople } from "#/endpoints/people";
import { keys } from "#/query-keys";
import {
	type ReminderRequest,
	type ReminderWithPerson,
	reminderRequestSchema,
} from "#/schemas/reminder";

interface ReminderFormProps {
	initial?: Partial<ReminderWithPerson>;
	onSubmit: (values: ReminderRequest) => Promise<void>;
	submitLabel?: string;
}

export function ReminderForm({
	initial,
	onSubmit,
	submitLabel = "Save Reminder",
}: ReminderFormProps) {
	const [apiError, setApiError] = useState<string | null>(null);

	const { data: peopleList } = useQuery({
		queryKey: keys.people.list({}),
		queryFn: () => listPeople({ page_size: 200 }),
	});

	const form = useForm({
		defaultValues: {
			title: initial?.title ?? "",
			notes: initial?.notes ?? "",
			due_date: initial?.due_date ?? "",
			person_id: initial?.person_id ?? null,
			important_date_id: initial?.important_date_id ?? null,
		} satisfies ReminderRequest,
		validators: {
			onSubmit: ({ value }) => {
				const result = reminderRequestSchema.safeParse(value);
				if (!result.success) {
					return result.error.issues.map((i) => i.message).join(", ");
				}
				return undefined;
			},
		},
		onSubmit: async ({ value }) => {
			setApiError(null);
			try {
				await onSubmit(value as ReminderRequest);
			} catch (err) {
				setApiError(
					err instanceof Error ? err.message : "Failed to save reminder",
				);
			}
		},
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

			<form.Field name="title">
				{(field) => (
					<FormField field={field} label="Title" placeholder="Call dentist" />
				)}
			</form.Field>

			<form.Field name="due_date">
				{(field) => <FormField field={field} label="Due Date" type="date" />}
			</form.Field>

			{/* Person picker (optional) */}
			<div className="space-y-1.5">
				<Label>Person (optional)</Label>
				<form.Field name="person_id">
					{(field) => (
						<Select
							value={field.state.value ? String(field.state.value) : ""}
							onValueChange={(v) => field.handleChange(v ? Number(v) : null)}
						>
							<SelectTrigger>
								<SelectValue placeholder="No person" />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="">No person</SelectItem>
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
	);
}
