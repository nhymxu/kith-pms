import { useForm } from "@tanstack/react-form";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { z } from "zod";
import { FormField } from "#/components/form/form-field";
import { SubmitButton } from "#/components/form/submit-button";
import { Alert, AlertDescription } from "#/components/ui/alert";
import { Label } from "#/components/ui/label";
import { Textarea } from "#/components/ui/textarea";
import { createPerson, updatePerson } from "#/endpoints/people";
import { keys } from "#/query-keys";
import {
	type Person,
	type PersonRequest,
	personRequestSchema,
} from "#/schemas/person";
import { PersonFormContacts } from "./person-form-contacts";
import { PersonFormLocations } from "./person-form-locations";

interface PersonFormProps {
	mode: "create" | "edit";
	initial?: Person;
}

export function PersonForm({ mode, initial }: PersonFormProps) {
	const navigate = useNavigate();
	const qc = useQueryClient();
	const [apiError, setApiError] = useState<string | null>(null);

	const form = useForm({
		defaultValues: {
			name: initial?.name ?? "",
			nickname: initial?.nickname ?? "",
			relationship_type: initial?.relationship_type ?? "",
			date_of_birth: initial?.date_of_birth ?? "",
			other_notes: initial?.other_notes ?? "",
			contacts: (initial?.contacts ?? []).map((c) => ({
				type: c.type,
				value: c.value,
				label: c.label,
				position: c.position,
			})),
			locations: (initial?.locations ?? []).map((l) => ({
				type: l.type,
				address: l.address,
				city: l.city,
				country: l.country,
				postal_code: l.postal_code,
				position: l.position,
			})),
		} satisfies PersonRequest,
		validators: {
			onSubmit: ({ value }) => {
				const r = personRequestSchema.safeParse(value);
				return r.success
					? undefined
					: r.error.issues.map((i) => i.message).join(", ");
			},
		},
		onSubmit: async ({ value }) => {
			setApiError(null);
			try {
				const parsed = personRequestSchema.parse(value);
				if (mode === "create") {
					const id = await createPerson(parsed);
					navigate({
						to: "/people/$personId",
						params: { personId: String(id) },
					});
				} else if (initial) {
					await updatePerson(initial.id, parsed);
					await qc.invalidateQueries({
						queryKey: keys.people.detail(initial.id),
					});
					await qc.invalidateQueries({ queryKey: keys.people.all });
					navigate({
						to: "/people/$personId",
						params: { personId: String(initial.id) },
					});
				}
			} catch (err) {
				if (err instanceof z.ZodError) {
					setApiError(err.issues.map((i) => i.message).join(", "));
				} else {
					setApiError(err instanceof Error ? err.message : "Save failed");
				}
			}
		},
	});

	return (
		<form
			onSubmit={(e) => {
				e.preventDefault();
				form.handleSubmit();
			}}
			className="space-y-6 max-w-2xl"
		>
			{apiError && (
				<Alert variant="destructive">
					<AlertDescription>{apiError}</AlertDescription>
				</Alert>
			)}

			{/* Identity */}
			<div className="space-y-4">
				<h2 className="text-sm font-medium uppercase tracking-wide text-zinc-500">
					Identity
				</h2>
				<div className="grid grid-cols-2 gap-4">
					<form.Field name="name">
						{(f) => (
							<FormField field={f} label="Name *" placeholder="Full name" />
						)}
					</form.Field>
					<form.Field name="nickname">
						{(f) => (
							<FormField field={f} label="Nickname" placeholder="Goes by…" />
						)}
					</form.Field>
				</div>
				<div className="grid grid-cols-2 gap-4">
					<form.Field name="relationship_type">
						{(f) => (
							<FormField
								field={f}
								label="Relationship type"
								placeholder="e.g. Friend"
							/>
						)}
					</form.Field>
					<form.Field name="date_of_birth">
						{(f) => <FormField field={f} label="Date of birth" type="date" />}
					</form.Field>
				</div>
				<form.Field name="other_notes">
					{(f) => (
						<div className="space-y-1.5">
							<Label htmlFor="other_notes">Notes</Label>
							<Textarea
								id="other_notes"
								value={f.state.value}
								onBlur={f.handleBlur}
								onChange={(e) => f.handleChange(e.target.value)}
								placeholder="Any other notes…"
								rows={3}
							/>
						</div>
					)}
				</form.Field>
			</div>

			{/* Contacts */}
			<form.Field name="contacts">
				{(f) => (
					<PersonFormContacts
						value={f.state.value}
						onChange={(rows) => f.handleChange(rows)}
					/>
				)}
			</form.Field>

			{/* Locations */}
			<form.Field name="locations">
				{(f) => (
					<PersonFormLocations
						value={f.state.value}
						onChange={(rows) => f.handleChange(rows)}
					/>
				)}
			</form.Field>

			<form.Subscribe selector={(s) => s.isSubmitting}>
				{(isSubmitting) => (
					<SubmitButton isPending={isSubmitting} pendingLabel="Saving…">
						{mode === "create" ? "Create person" : "Save changes"}
					</SubmitButton>
				)}
			</form.Subscribe>
		</form>
	);
}
