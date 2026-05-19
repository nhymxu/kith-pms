// Gift create/edit form — TanStack Form + Zod validation
import { useForm } from "@tanstack/react-form";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { FormField } from "#/components/form/form-field";
import { SubmitButton } from "#/components/form/submit-button";
import { Alert, AlertDescription } from "#/components/ui/alert";
import { Button } from "#/components/ui/button";
import { Input } from "#/components/ui/input";
import { Label } from "#/components/ui/label";
import { Textarea } from "#/components/ui/textarea";
import { listPeople } from "#/endpoints/people";
import { keys } from "#/query-keys";
import {
	type GiftRequest,
	type GiftWithPerson,
	giftRequestSchema,
} from "#/schemas/gift";

interface GiftFormProps {
	initial?: Partial<GiftWithPerson>;
	onSubmit: (values: GiftRequest) => Promise<void>;
	submitLabel?: string;
	onCancel?: () => void;
	onImageChange?: (file: File | null) => void;
}

export function GiftForm({
	initial,
	onSubmit,
	submitLabel = "Save Gift",
	onCancel,
	onImageChange,
}: GiftFormProps) {
	const [apiError, setApiError] = useState<string | null>(null);
	const [personSearch, setPersonSearch] = useState("");
	const [amountDisplay, setAmountDisplay] = useState(
		initial?.amount_cents != null
			? (initial.amount_cents / 100).toString()
			: "",
	);

	const { data: peopleList } = useQuery({
		queryKey: keys.people.list({}),
		queryFn: () => listPeople({ page_size: 200 }),
	});

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
				const result = giftRequestSchema.safeParse(value);
				if (!result.success) {
					return result.error.issues.map((i) => i.message).join(", ");
				}
				return undefined;
			},
		},
		onSubmit: async ({ value }) => {
			setApiError(null);
			try {
				await onSubmit(value as GiftRequest);
			} catch (err) {
				setApiError(err instanceof Error ? err.message : "Failed to save gift");
			}
		},
	});

	const filteredPeople =
		peopleList?.items.filter((p) =>
			p.name.toLowerCase().includes(personSearch.toLowerCase()),
		) ?? [];

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

			{/* Person search combobox */}
			<div className="space-y-1.5">
				<Label>Person</Label>
				<form.Field name="person_id">
					{(field) => {
						const selected = peopleList?.items.find(
							(p) => p.id === field.state.value,
						);
						return (
							<div className="space-y-1">
								<Input
									placeholder="Search person…"
									value={
										personSearch !== ""
											? personSearch
											: selected
												? selected.name
												: ""
									}
									onChange={(e) => setPersonSearch(e.target.value)}
									onFocus={() => setPersonSearch("")}
								/>
								{personSearch && (
									<div className="border border-zinc-200 rounded-md bg-white shadow-sm max-h-48 overflow-y-auto">
										{filteredPeople.length === 0 ? (
											<p className="px-3 py-2 text-sm text-zinc-400">
												No results
											</p>
										) : (
											filteredPeople.map((p) => (
												<button
													key={p.id}
													type="button"
													className="w-full text-left px-3 py-2 text-sm hover:bg-zinc-50"
													onClick={() => {
														field.handleChange(p.id);
														setPersonSearch("");
													}}
												>
													{p.name}
												</button>
											))
										)}
									</div>
								)}
							</div>
						);
					}}
				</form.Field>
			</div>

			<form.Field name="title">
				{(field) => (
					<FormField
						field={field}
						label="Title / Occasion"
						placeholder="Birthday gift"
					/>
				)}
			</form.Field>

			{/* Direction radio group */}
			<div className="space-y-1.5">
				<Label>Direction</Label>
				<form.Field name="direction">
					{(field) => (
						<div className="flex gap-4">
							{(["given", "received", "planned"] as const).map((v) => (
								<label
									key={v}
									className="flex items-center gap-1.5 cursor-pointer"
								>
									<input
										type="radio"
										name="direction"
										value={v}
										checked={field.state.value === v}
										onChange={() => field.handleChange(v)}
									/>
									<span className="capitalize text-sm">{v}</span>
								</label>
							))}
						</div>
					)}
				</form.Field>
			</div>

			{/* Debt type radio group */}
			<div className="space-y-1.5">
				<Label>Debt type</Label>
				<form.Field name="debt_type">
					{(field) => (
						<div className="flex gap-4">
							{(
								[
									["", "None"],
									["i_owe", "I owe"],
									["they_owe", "They owe"],
								] as const
							).map(([v, label]) => (
								<label
									key={v}
									className="flex items-center gap-1.5 cursor-pointer"
								>
									<input
										type="radio"
										name="debt_type"
										value={v}
										checked={(field.state.value ?? "") === v}
										onChange={() => field.handleChange(v)}
									/>
									<span className="text-sm">{label}</span>
								</label>
							))}
						</div>
					)}
				</form.Field>
			</div>

			{/* Amount + currency */}
			<div className="flex gap-3">
				<div className="flex-1 space-y-1.5">
					<Label>Amount</Label>
					<form.Field name="amount_cents">
						{(field) => (
							<input
								type="number"
								min="0"
								step="0.01"
								placeholder="0.00"
								value={amountDisplay}
								className="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
								onChange={(e) => {
									setAmountDisplay(e.target.value);
									const cents = Math.round(
										parseFloat(e.target.value || "0") * 100,
									);
									field.handleChange(isNaN(cents) ? null : cents);
								}}
								onBlur={field.handleBlur}
							/>
						)}
					</form.Field>
				</div>
				<div className="w-24 space-y-1.5">
					<form.Field name="currency">
						{(field) => (
							<FormField
								field={field}
								label="Currency"
								placeholder="USD"
								maxLength={3}
							/>
						)}
					</form.Field>
				</div>
			</div>

			<form.Field name="date">
				{(field) => <FormField field={field} label="Date" type="date" />}
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

			{onImageChange && (
				<div className="space-y-1.5">
					<Label>Image</Label>
					<input
						type="file"
						accept="image/jpeg,image/png,image/gif,image/webp"
						onChange={(e) => onImageChange(e.target.files?.[0] ?? null)}
						className="text-sm"
					/>
				</div>
			)}

			<form.Subscribe selector={(s) => s.isSubmitting}>
				{(isSubmitting) => (
					<div className="flex gap-2">
						{onCancel && (
							<Button
								type="button"
								variant="neutral"
								onClick={onCancel}
								disabled={isSubmitting}
							>
								Cancel
							</Button>
						)}
						<SubmitButton isPending={isSubmitting} pendingLabel="Saving…">
							{submitLabel}
						</SubmitButton>
					</div>
				)}
			</form.Subscribe>
		</form>
	);
}
