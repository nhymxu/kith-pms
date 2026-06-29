// Reminder create/edit form — TanStack Form + Zod validation
import { useForm } from "@tanstack/react-form";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { FormField } from "#/components/form/form-field";
import { SubmitButton } from "#/components/form/submit-button";
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

const DAY_NAMES = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];

const DAYS_BEFORE_OPTIONS = [
	{ value: 0, label: "On the day" },
	{ value: 1, label: "1 day before" },
	{ value: 3, label: "3 days before" },
	{ value: 7, label: "1 week before" },
	{ value: 14, label: "2 weeks before" },
	{ value: 30, label: "1 month before" },
];

function recurrenceLabel(rule: RecurrenceRule): string {
	switch (rule.type) {
		case "daily":
			return "Daily";
		case "weekly":
			return "Weekly";
		case "monthly":
			return "Monthly";
		case "yearly":
			return "Yearly";
		case "custom": {
			const n = rule.interval ?? 1;
			const u = rule.unit ?? "days";
			return `Every ${n} ${u}`;
		}
		case "day_of_week":
			return `Every ${DAY_NAMES[rule.day_of_week ?? 0]}`;
		case "relative_contact":
			return `${rule.interval ?? 30}d after contact`;
		default:
			return "Recurring";
	}
}

export function ReminderForm({
	initial,
	onSubmit,
	submitLabel = "Save Reminder",
}: ReminderFormProps) {
	const [apiError, setApiError] = useState<string | null>(null);
	const isBirthdayInitial = initial?.recurrence_rule?.type === "birthday";
	const [isBirthday, setIsBirthday] = useState(isBirthdayInitial);
	const [recurring, setRecurring] = useState(
		!!initial?.recurrence_rule && !isBirthdayInitial,
	);
	const [daysBefore, setDaysBefore] = useState(
		initial?.recurrence_rule?.days_before_dob ?? 0,
	);

	const { data: peopleList } = useQuery({
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

			{!isBirthday && (
				<form.Field name="due_date">
					{(field) => (
						<FormField field={field} label="Due Date" type="datetime-local" />
					)}
				</form.Field>
			)}

			{/* Person picker (optional) */}
			<div className="space-y-1.5">
				<Label>Person (optional)</Label>
				<form.Field name="person_id">
					{(field) => (
						<Select
							value={field.state.value ? String(field.state.value) : "__none__"}
							onValueChange={(v) =>
								field.handleChange(v === "__none__" ? null : Number(v))
							}
						>
							<SelectTrigger>
								<SelectValue placeholder="No person" />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="__none__">No person</SelectItem>
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
				<Label htmlFor="birthday-toggle">Birthday reminder</Label>
			</div>

			{isBirthday && (
				<div className="space-y-3 rounded-md border border-zinc-200 p-3">
					<p className="text-sm text-zinc-500">
						Select a person with a date of birth. Due date is computed
						automatically.
					</p>
					<div className="space-y-1.5">
						<Label>Remind me</Label>
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
										{opt.label}
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
					<Label htmlFor="recurring-toggle">Recurring</Label>
				</div>
			)}

			{recurring && (
				<div className="space-y-3 rounded-md border border-zinc-200 p-3">
					<form.Field name="recurrence_rule">
						{(field) => {
							const rule = field.state.value ?? { type: "weekly" as const };
							return (
								<div className="space-y-3">
									<div className="space-y-1.5">
										<Label>Repeat</Label>
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
												<SelectItem value="daily">Daily</SelectItem>
												<SelectItem value="weekly">Weekly</SelectItem>
												<SelectItem value="monthly">Monthly</SelectItem>
												<SelectItem value="yearly">Yearly</SelectItem>
												<SelectItem value="custom">Custom interval</SelectItem>
												<SelectItem value="day_of_week">Day of week</SelectItem>
												<SelectItem value="relative_contact">
													Relative to last contact
												</SelectItem>
											</SelectContent>
										</Select>
									</div>

									{rule.type === "custom" && (
										<div className="flex items-center gap-2">
											<Label className="shrink-0">Every</Label>
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
													<SelectItem value="days">days</SelectItem>
													<SelectItem value="weeks">weeks</SelectItem>
													<SelectItem value="months">months</SelectItem>
												</SelectContent>
											</Select>
										</div>
									)}

									{rule.type === "day_of_week" && (
										<div className="space-y-1.5">
											<Label>Day</Label>
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
													{DAY_NAMES.map((d, i) => (
														<SelectItem key={d} value={String(i)}>
															{d}
														</SelectItem>
													))}
												</SelectContent>
											</Select>
										</div>
									)}

									{rule.type === "relative_contact" && (
										<div className="flex items-center gap-2">
											<Label className="shrink-0">Every</Label>
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
											<span className="text-sm text-zinc-500">
												days after last journal entry
											</span>
										</div>
									)}
								</div>
							);
						}}
					</form.Field>

					<form.Field name="recurrence_end_date">
						{(field) => (
							<FormField
								field={field}
								label="End date (optional)"
								type="date"
							/>
						)}
					</form.Field>
				</div>
			)}

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

export { recurrenceLabel };
