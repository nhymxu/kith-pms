// Gift create/edit form — TanStack Form + Zod validation

import { useForm } from "@tanstack/react-form";
import { useSuspenseQuery } from "@tanstack/react-query";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { FormField } from "#/components/form/form-field";
import { SubmitButton } from "#/components/form/submit-button";
import { QueryBoundary } from "#/components/query-boundary";
import { Alert, AlertDescription } from "#/components/ui/alert";
import { Button } from "#/components/ui/button";
import { Input } from "#/components/ui/input";
import { Label } from "#/components/ui/label";
import { Textarea } from "#/components/ui/textarea";
import { listPeople } from "#/endpoints/people";
import { GiftImageUploader } from "#/features/gifts/gift-image-uploader";
import { apiErrorMessage } from "#/lib/api-error-message";
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

function FormFallback() {
	const { t } = useTranslation("gifts");
	return (
		<div className="py-8 text-center text-[13px] text-sub">
			{t("form.loading")}
		</div>
	);
}

export function GiftForm(props: GiftFormProps) {
	return (
		<QueryBoundary fallback={<FormFallback />}>
			<GiftFormInner {...props} />
		</QueryBoundary>
	);
}

function GiftFormInner({
	initial,
	onSubmit,
	submitLabel,
	onCancel,
	onImageChange,
}: GiftFormProps) {
	const { t } = useTranslation("gifts");
	const [apiError, setApiError] = useState<string | null>(null);
	const [personSearch, setPersonSearch] = useState("");
	const [croppedPreview, setCroppedPreview] = useState<string | null>(null);
	const croppedPreviewRef = useRef<string | null>(null);
	croppedPreviewRef.current = croppedPreview;
	useEffect(() => {
		return () => {
			if (croppedPreviewRef.current) {
				URL.revokeObjectURL(croppedPreviewRef.current);
			}
		};
	}, []);
	const [amountDisplay, setAmountDisplay] = useState(
		initial?.amount_cents != null
			? (initial.amount_cents / 100).toString()
			: "",
	);

	const { data: peopleList } = useSuspenseQuery({
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
				setApiError(apiErrorMessage(err, t("form.saveFailed")));
			}
		},
	});

	const filteredPeople = peopleList.items.filter((p) =>
		p.name.toLowerCase().includes(personSearch.toLowerCase()),
	);

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
				<Label>{t("form.person")}</Label>
				<form.Field name="person_id">
					{(field) => {
						const selected = peopleList.items.find(
							(p) => p.id === field.state.value,
						);
						return (
							<div className="space-y-1">
								<Input
									placeholder={t("form.searchPersonPlaceholder")}
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
									<div className="border-bw border-line rounded-base bg-panel shadow-shadow max-h-48 overflow-y-auto">
										{filteredPeople.length === 0 ? (
											<p className="px-3 py-2 text-sm text-sub">
												{t("form.noResults")}
											</p>
										) : (
											filteredPeople.map((p) => (
												<button
													key={p.id}
													type="button"
													className="w-full text-left px-3 py-2 text-sm hover:bg-chip"
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
						label={t("form.titleLabel")}
						placeholder={t("form.titlePlaceholder")}
					/>
				)}
			</form.Field>

			{/* Direction radio group */}
			<div className="space-y-1.5">
				<Label>{t("form.direction")}</Label>
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
									<span className="capitalize text-sm">
										{t(`direction.${v}`)}
									</span>
								</label>
							))}
						</div>
					)}
				</form.Field>
			</div>

			{/* Debt type radio group */}
			<div className="space-y-1.5">
				<Label>{t("form.debtType")}</Label>
				<form.Field name="debt_type">
					{(field) => (
						<div className="flex gap-4">
							{(
								[
									["", t("debt.none")],
									["i_owe", t("debt.iOwe")],
									["they_owe", t("debt.theyOwe")],
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
					<Label>{t("form.amount")}</Label>
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
									field.handleChange(Number.isNaN(cents) ? null : cents);
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
								label={t("form.currency")}
								placeholder="USD"
								maxLength={3}
							/>
						)}
					</form.Field>
				</div>
			</div>

			<form.Field name="date">
				{(field) => (
					<FormField field={field} label={t("form.date")} type="date" />
				)}
			</form.Field>

			{/* Notes */}
			<div className="space-y-1.5">
				<Label>{t("form.notes")}</Label>
				<form.Field name="notes">
					{(field) => (
						<Textarea
							value={field.state.value}
							onBlur={field.handleBlur}
							onChange={(e) => field.handleChange(e.target.value)}
							placeholder={t("form.notesPlaceholder")}
							rows={3}
						/>
					)}
				</form.Field>
			</div>

			{onImageChange && (
				<div className="space-y-1.5">
					<Label>{t("form.image")}</Label>
					<GiftImageUploader
						imageSrc={croppedPreview}
						hasImage={Boolean(croppedPreview)}
						onUpload={(file) => {
							const nextPreview = URL.createObjectURL(file);
							setCroppedPreview((prev) => {
								if (prev) URL.revokeObjectURL(prev);
								return nextPreview;
							});
							onImageChange(file);
						}}
						onRemove={() => {
							if (croppedPreview) URL.revokeObjectURL(croppedPreview);
							setCroppedPreview(null);
							onImageChange(null);
						}}
						uploadPending={false}
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
								{t("form.cancel")}
							</Button>
						)}
						<SubmitButton
							isPending={isSubmitting}
							pendingLabel={t("form.saving")}
						>
							{submitLabel ?? t("form.save")}
						</SubmitButton>
					</div>
				)}
			</form.Subscribe>
		</form>
	);
}
