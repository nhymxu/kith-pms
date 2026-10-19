// Reminder create/edit form — TanStack Form + Zod validation

import { useForm } from "@tanstack/react-form";
import { useSuspenseQuery } from "@tanstack/react-query";
import type { TFunction } from "i18next";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { FormField } from "#/components/form/form-field";
import { SubmitButton } from "#/components/form/submit-button";
import { QueryBoundary } from "#/components/query-boundary";
import { Alert, AlertDescription } from "#/components/ui/alert";
import { Input } from "#/components/ui/input";
import { Label } from "#/components/ui/label";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "#/components/ui/select";
import { Switch } from "#/components/ui/switch";
import { Textarea } from "#/components/ui/textarea";
import { listPeople } from "#/endpoints/people";
import { apiErrorMessage } from "#/lib/api-error-message";
import { datetimeLocalToUtc, utcToDatetimeLocal } from "#/lib/format-datetime";
import { keys } from "#/query-keys";
import {
	type RecurrenceRule,
	type ReminderRequest,
	type ReminderWithPerson,
	reminderRequestSchema,
} from "#/schemas/reminder";

interface ReminderFormProps {
	initial?: Partial<ReminderWithPerson>;
	onSubmit: (values: ReminderRequest) => Promise<void>;
	submitLabel?: string;
}

const DAY_KEYS = ["sun", "mon", "tue", "wed", "thu", "fri", "sat"] as const;

const DAYS_BEFORE_OPTIONS = [
	{ value: 0, key: "onTheDay" },
	{ value: 1, key: "oneDay" },
	{ value: 3, key: "threeDays" },
	{ value: 7, key: "oneWeek" },
	{ value: 14, key: "twoWeeks" },
	{ value: 30, key: "oneMonth" },
] as const;

export function recurrenceLabel(
	rule: RecurrenceRule,
	t: TFunction<"reminders">,
): string {
	switch (rule.type) {
		case "daily":
			return t("recurrence.daily");
		case "weekly":
			return t("recurrence.weekly");
		case "monthly":
			return t("recurrence.monthly");
		case "yearly":
			return t("recurrence.yearly");
		case "custom": {
			const count = rule.interval ?? 1;
			switch (rule.unit ?? "days") {
				case "weeks":
					return t("recurrence.everyWeeks", { count });
				case "months":
					return t("recurrence.everyMonths", { count });
				default:
					return t("recurrence.everyDays", { count });
			}
		}
		case "day_of_week":
			return t("recurrence.everyDay", {
				day: t(`days.${DAY_KEYS[rule.day_of_week ?? 0]}`),
			});
		case "relative_contact":
			return t("recurrence.afterContact", { interval: rule.interval ?? 30 });
		default:
			return t("recurrence.recurring");
	}
}

function FormFallback() {
	const { t } = useTranslation("reminders");
	return (
		<div className="py-8 text-center text-[13px] text-sub">
			{t("form.loading")}
		</div>
	);
}

export function ReminderForm(props: ReminderFormProps) {
	return (
		<QueryBoundary fallback={<FormFallback />}>
			<ReminderFormInner {...props} />
		</QueryBoundary>
	);
}

function ReminderFormInner({
	initial,
	onSubmit,
	submitLabel,
}: ReminderFormProps) {
	const { t } = useTranslation("reminders");
	const [apiError, setApiError] = useState<string | null>(null);
	const isBirthdayInitial = initial?.recurrence_rule?.type === "birthday";
	const [isBirthday, setIsBirthday] = useState(isBirthdayInitial);
	const [recurring, setRecurring] = useState(
		!!initial?.recurrence_rule && !isBirthdayInitial,
	);
	const [daysBefore, setDaysBefore] = useState(
		initial?.recurrence_rule?.days_before_dob ?? 0,
	);

	const { data: peopleList } = useSuspenseQuery({
		queryKey: keys.people.list({}),
		queryFn: () => listPeople({ page_size: 200 }),
	});

	const form = useForm({
		defaultValues: {
			title: initial?.title ?? "",
			notes: initial?.notes ?? "",
			due_date: utcToDatetimeLocal(initial?.due_date ?? ""),
			person_id: initial?.person_id ?? null,
			important_date_id: initial?.important_date_id ?? null,
			recurrence_rule: initial?.recurrence_rule ?? null,
			recurrence_end_date: initial?.recurrence_end_date ?? null,
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
				let payload: ReminderRequest;
				if (isBirthday) {
					payload = {
						...value,
						due_date: "",
						recurrence_rule: {
							type: "birthday",
							days_before_dob: daysBefore,
						},
						recurrence_end_date: null,
					};
				} else {
					payload = {
						...value,
						due_date: datetimeLocalToUtc(value.due_date),
						recurrence_rule: recurring ? value.recurrence_rule : null,
						recurrence_end_date: recurring ? value.recurrence_end_date : null,
					};
				}
				await onSubmit(payload as ReminderRequest);
			} catch (err) {
				setApiError(apiErrorMessage(err, t("form.saveFailed")));
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
					<FormField
						field={field}
						label={t("form.title")}
						placeholder={t("form.titlePlaceholder")}
					/>
				)}
			</form.Field>

			{!isBirthday && (
				<form.Field name="due_date">
					{(field) => (
						<FormField
							field={field}
							label={t("form.dueDate")}
							type="datetime-local"
						/>
					)}
				</form.Field>
			)}

			{/* Person picker (optional) */}
			<div className="space-y-1.5">
				<Label>{t("form.person")}</Label>
				<form.Field name="person_id">
					{(field) => (
						<Select
							value={field.state.value ? String(field.state.value) : "__none__"}
							onValueChange={(v) =>
								field.handleChange(v === "__none__" ? null : Number(v))
							}
						>
							<SelectTrigger>
								<SelectValue placeholder={t("form.noPerson")} />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="__none__">{t("form.noPerson")}</SelectItem>
								{peopleList.items.map((p) => (
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

			{/* Birthday reminder toggle */}
			<div className="flex items-center gap-2">
				<Switch
					id="birthday-toggle"
					checked={isBirthday}
					onCheckedChange={(checked) => {
						setIsBirthday(checked);
						if (checked) {
							setRecurring(false);
							form.setFieldValue("recurrence_rule", null);
							form.setFieldValue("recurrence_end_date", null);
						}
					}}
				/>
				<Label htmlFor="birthday-toggle">{t("form.birthdayReminder")}</Label>
			</div>

			{isBirthday && (
				<div className="space-y-3 rounded-base border-bw border-line p-3">
					<p className="text-sm text-sub">{t("form.birthdayHint")}</p>
					<div className="space-y-1.5">
						<Label>{t("form.remindMe")}</Label>
						<Select
							value={String(daysBefore)}
							onValueChange={(v) => setDaysBefore(Number(v))}
						>
							<SelectTrigger>
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								{DAYS_BEFORE_OPTIONS.map((opt) => (
									<SelectItem key={opt.value} value={String(opt.value)}>
										{t(`daysBefore.${opt.key}`)}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</div>
				</div>
			)}

			{/* Recurrence toggle — mutually exclusive with birthday */}
			{!isBirthday && (
				<div className="flex items-center gap-2">
					<Switch
						id="recurring-toggle"
						checked={recurring}
						onCheckedChange={(checked) => {
							setRecurring(checked);
							if (!checked) {
								form.setFieldValue("recurrence_rule", null);
								form.setFieldValue("recurrence_end_date", null);
							} else {
								form.setFieldValue("recurrence_rule", { type: "weekly" });
							}
						}}
					/>
					<Label htmlFor="recurring-toggle">{t("form.recurring")}</Label>
				</div>
			)}

			{recurring && (
				<div className="space-y-3 rounded-base border-bw border-line p-3">
					<form.Field name="recurrence_rule">
						{(field) => {
							const rule = field.state.value ?? { type: "weekly" as const };
							return (
								<div className="space-y-3">
									<div className="space-y-1.5">
										<Label>{t("form.repeat")}</Label>
										<Select
											value={rule.type}
											onValueChange={(v) =>
												field.handleChange({
													...rule,
													type: v as RecurrenceRule["type"],
												})
											}
										>
											<SelectTrigger>
												<SelectValue />
											</SelectTrigger>
											<SelectContent>
												<SelectItem value="daily">
													{t("recurrence.daily")}
												</SelectItem>
												<SelectItem value="weekly">
													{t("recurrence.weekly")}
												</SelectItem>
												<SelectItem value="monthly">
													{t("recurrence.monthly")}
												</SelectItem>
												<SelectItem value="yearly">
													{t("recurrence.yearly")}
												</SelectItem>
												<SelectItem value="custom">
													{t("recurrence.customInterval")}
												</SelectItem>
												<SelectItem value="day_of_week">
													{t("recurrence.dayOfWeek")}
												</SelectItem>
												<SelectItem value="relative_contact">
													{t("recurrence.relativeContact")}
												</SelectItem>
											</SelectContent>
										</Select>
									</div>

									{rule.type === "custom" && (
										<div className="flex items-center gap-2">
											<Label className="shrink-0">{t("form.every")}</Label>
											<Input
												type="number"
												min={1}
												className="w-20"
												value={rule.interval ?? 1}
												onChange={(e) =>
													field.handleChange({
														...rule,
														interval: Number(e.target.value),
													})
												}
											/>
											<Select
												value={rule.unit ?? "days"}
												onValueChange={(v) =>
													field.handleChange({
														...rule,
														unit: v as RecurrenceRule["unit"],
													})
												}
											>
												<SelectTrigger className="w-28">
													<SelectValue />
												</SelectTrigger>
												<SelectContent>
													<SelectItem value="days">
														{t("units.days")}
													</SelectItem>
													<SelectItem value="weeks">
														{t("units.weeks")}
													</SelectItem>
													<SelectItem value="months">
														{t("units.months")}
													</SelectItem>
												</SelectContent>
											</Select>
										</div>
									)}

									{rule.type === "day_of_week" && (
										<div className="space-y-1.5">
											<Label>{t("form.day")}</Label>
											<Select
												value={String(rule.day_of_week ?? 1)}
												onValueChange={(v) =>
													field.handleChange({
														...rule,
														day_of_week: Number(v),
													})
												}
											>
												<SelectTrigger>
													<SelectValue />
												</SelectTrigger>
												<SelectContent>
													{DAY_KEYS.map((d, i) => (
														<SelectItem key={d} value={String(i)}>
															{t(`days.${d}`)}
														</SelectItem>
													))}
												</SelectContent>
											</Select>
										</div>
									)}

									{rule.type === "relative_contact" && (
										<div className="flex items-center gap-2">
											<Label className="shrink-0">{t("form.every")}</Label>
											<Input
												type="number"
												min={1}
												className="w-20"
												value={rule.interval ?? 30}
												onChange={(e) =>
													field.handleChange({
														...rule,
														interval: Number(e.target.value),
													})
												}
											/>
											<span className="text-sm text-sub">
												{t("form.daysAfterEntry")}
											</span>
										</div>
									)}
								</div>
							);
						}}
					</form.Field>

					<form.Field name="recurrence_end_date">
						{(field) => (
							<FormField field={field} label={t("form.endDate")} type="date" />
						)}
					</form.Field>
				</div>
			)}

			<form.Subscribe selector={(s) => s.isSubmitting}>
				{(isSubmitting) => (
					<SubmitButton
						isPending={isSubmitting}
						pendingLabel={t("form.saving")}
					>
						{submitLabel ?? t("form.save")}
					</SubmitButton>
				)}
			</form.Subscribe>
		</form>
	);
}
