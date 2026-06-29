import { useForm } from "@tanstack/react-form";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { X } from "lucide-react";
import { useEffect, useState } from "react";
import { z } from "zod";
import { FormField } from "#/components/form/form-field";
import { SubmitButton } from "#/components/form/submit-button";
import { Alert, AlertDescription } from "#/components/ui/alert";
import { Button } from "#/components/ui/button";
import { Checkbox } from "#/components/ui/checkbox";
import { Label } from "#/components/ui/label";
import { RadioGroup, RadioGroupItem } from "#/components/ui/radio-group";
import { Textarea } from "#/components/ui/textarea";
import { createPerson, updatePerson } from "#/endpoints/people";
import { datetimeLocalToUtc, utcToDatetimeLocal } from "#/lib/format-datetime";
import { keys } from "#/query-keys";
import {
	genderOptions,
	type Person,
	type PersonRequest,
	personRequestSchema,
} from "#/schemas/person";
import { PersonFormContacts } from "./person-form-contacts";
import { PersonFormLocations } from "./person-form-locations";

// Extracted to avoid calling form field setters inside a Subscribe render callback.
function BirthdayCheckbox({
	checked,
	onChange,
	dobSet,
	onDobClear,
}: {
	checked: boolean;
	onChange: (v: boolean) => void;
	dobSet: boolean;
	onDobClear: () => void;
}) {
	useEffect(() => {
		if (!dobSet && checked) onDobClear();
	}, [dobSet, checked, onDobClear]);

	if (!dobSet) return null;

	return (
		<div className="flex items-center gap-2">
			<Checkbox
				id="create_birthday_reminder"
				checked={checked}
				onCheckedChange={(v) => onChange(v === true)}
			/>
			<Label
				htmlFor="create_birthday_reminder"
				className="font-normal cursor-pointer"
			>
				Create an annual birthday reminder
			</Label>
		</div>
	);
}

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
			gender: initial?.gender ?? "",
			relationship_type: initial?.relationship_type ?? "",
			date_of_birth: initial?.date_of_birth ?? "",
			create_birthday_reminder: initial?.has_birthday_reminder ?? false,
			last_contact_at: utcToDatetimeLocal(initial?.last_contact_at),
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
				const parsed = personRequestSchema.parse({
					...value,
					last_contact_at: value.last_contact_at
						? datetimeLocalToUtc(value.last_contact_at)
						: null,
				});
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

				{/* Birthday reminder checkbox — only shown when DOB is set */}
				<form.Subscribe
					selector={(s) => ({
						dob: s.values.date_of_birth,
						checked: s.values.create_birthday_reminder,
					})}
				>
					{({ dob, checked }) => (
						<BirthdayCheckbox
							checked={checked}
							onChange={(v) =>
								form.setFieldValue("create_birthday_reminder", v)
							}
							dobSet={!!dob}
							onDobClear={() =>
								form.setFieldValue("create_birthday_reminder", false)
							}
						/>
					)}
				</form.Subscribe>

				<form.Field name="last_contact_at">
					{(f) => (
						<div>
							<FormField field={f} label="Last contact" type="datetime-local" />
							{f.state.value && (
								<Button
									type="button"
									variant="neutral"
									size="sm"
									className="mt-1.5 h-6 px-2 text-xs"
									onClick={() => f.handleChange("")}
								>
									<X className="size-3" /> Clear last contact
								</Button>
							)}
						</div>
					)}
				</form.Field>
				<form.Field name="gender">
					{(f) => (
						<div className="space-y-1.5">
							<Label>Gender</Label>
							<RadioGroup
								value={f.state.value}
								onValueChange={(v: string) => f.handleChange(v)}
								className="flex flex-wrap gap-4"
							>
								<div className="flex items-center gap-2">
									<RadioGroupItem value="" id="gender-unselected" />
									<Label
										htmlFor="gender-unselected"
										className="font-normal cursor-pointer text-zinc-400"
									>
										Unselected
									</Label>
								</div>
								{genderOptions.map((opt) => (
									<div key={opt.value} className="flex items-center gap-2">
										<RadioGroupItem
											value={opt.value}
											id={`gender-${opt.value}`}
										/>
										<Label
											htmlFor={`gender-${opt.value}`}
											className="font-normal cursor-pointer"
										>
											{opt.label}
										</Label>
									</div>
								))}
							</RadioGroup>
						</div>
					)}
				</form.Field>
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
