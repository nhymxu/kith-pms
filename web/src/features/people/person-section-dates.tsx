import {
	useMutation,
	useQueryClient,
	useSuspenseQuery,
} from "@tanstack/react-query";
import { Check, Lock, Pencil, Plus, Trash2, X } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { QueryBoundary } from "#/components/query-boundary";
import { SafeTrans } from "#/components/safe-trans";
import { Alert, AlertDescription } from "#/components/ui/alert";
import { Badge } from "#/components/ui/badge";
import { Button } from "#/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import { Input } from "#/components/ui/input";
import {
	listDatesByPerson,
	replaceDatesForPerson,
} from "#/endpoints/important-dates";
import { apiErrorMessage } from "#/lib/api-error-message";
import { keys } from "#/query-keys";
import type { ImportantDate } from "#/schemas/date";
import type { Person } from "#/schemas/person";

function SectionHeading({ children }: { children: React.ReactNode }) {
	return (
		<h2 className="text-[11px] font-semibold uppercase tracking-widest text-sub mb-2">
			{children}
		</h2>
	);
}

const DATE_KINDS = ["Anniversary", "Memorial", "Custom"] as const;

interface DateFormProps {
	date: Partial<ImportantDate>;
	onSave: (d: Partial<ImportantDate>) => void;
	onCancel: () => void;
}

function DateForm({ date, onSave, onCancel }: DateFormProps) {
	const { t } = useTranslation("people");
	const [kind, setKind] = useState(date.kind ?? "Custom");
	const [label, setLabel] = useState(date.label ?? "");
	const [dateValue, setDateValue] = useState(date.date_value ?? "");
	const [recurring, setRecurring] = useState(date.recurring ?? false);

	return (
		<div className="border-bw border-line rounded-md p-3 bg-muted space-y-2">
			<div className="flex gap-2">
				<select
					className="h-8 border-field-bw border-field-line rounded-md bg-field px-2 text-sm"
					value={kind}
					onChange={(e) => setKind(e.target.value)}
				>
					{DATE_KINDS.map((k) => (
						<option key={k} value={k}>
							{t(`sections.dates.kind.${k}`)}
						</option>
					))}
				</select>
				<Input
					className="h-8 flex-1"
					value={label}
					onChange={(e) => setLabel(e.target.value)}
					placeholder={t("sections.dates.labelPlaceholder")}
				/>
			</div>
			<div className="flex gap-2 items-center">
				<Input
					className="h-8 flex-1"
					value={dateValue}
					onChange={(e) => setDateValue(e.target.value)}
					placeholder={t("sections.dates.datePlaceholder")}
				/>
				<label className="flex items-center gap-1 text-sm text-sub cursor-pointer">
					<input
						type="checkbox"
						checked={recurring}
						onChange={(e) => setRecurring(e.target.checked)}
					/>
					{t("sections.dates.recurring")}
				</label>
			</div>
			<div className="flex gap-2 justify-end">
				<Button variant="neutral" size="sm" onClick={onCancel}>
					<X className="size-3" /> {t("sections.dates.cancel")}
				</Button>
				<Button
					size="sm"
					disabled={!dateValue}
					onClick={() =>
						onSave({ ...date, kind, label, date_value: dateValue, recurring })
					}
				>
					<Check className="size-3" /> {t("sections.dates.save")}
				</Button>
			</div>
		</div>
	);
}

interface ImportantDatesSectionInnerProps {
	personId: number;
	person: Person;
}

function ImportantDatesSectionInner({
	personId,
	person,
}: ImportantDatesSectionInnerProps) {
	const { t } = useTranslation("people");
	const qc = useQueryClient();
	const [editingId, setEditingId] = useState<number | "new" | null>(null);
	const [saveError, setSaveError] = useState<string | null>(null);
	const [confirmDeleteId, setConfirmDeleteId] = useState<number | null>(null);

	const { data } = useSuspenseQuery({
		queryKey: keys.dates.list(personId),
		queryFn: () => listDatesByPerson(personId),
	});

	const saveMutation = useMutation({
		mutationFn: (dates: ImportantDate[]) =>
			replaceDatesForPerson(personId, {
				dates: dates.map((d, i) => ({
					kind: d.kind,
					label: d.label ?? "",
					date_value: d.date_value,
					recurring: d.recurring ?? false,
					notes: d.notes ?? "",
					position: i,
				})),
			}),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.dates.list(personId) });
			setEditingId(null);
			setSaveError(null);
		},
		onError: (e) =>
			setSaveError(apiErrorMessage(e, t("sections.dates.saveFailed"))),
	});

	function handleSave(updated: Partial<ImportantDate>) {
		let dates: ImportantDate[];
		if (editingId === "new") {
			dates = [
				...data,
				{
					id: 0,
					person_id: personId,
					created_at: "",
					position: data.length,
					kind: updated.kind ?? "",
					label: updated.label ?? "",
					date_value: updated.date_value ?? "",
					recurring: updated.recurring ?? false,
					notes: "",
				},
			];
		} else {
			dates = data.map((d) => (d.id === editingId ? { ...d, ...updated } : d));
		}
		saveMutation.mutate(dates);
	}

	function handleDelete(id: number) {
		saveMutation.mutate(data.filter((d) => d.id !== id));
	}

	return (
		<div>
			<div className="flex items-center justify-between mb-2">
				<SectionHeading>{t("sections.dates.title")}</SectionHeading>
				<Button variant="neutral" size="sm" onClick={() => setEditingId("new")}>
					<Plus className="size-3" /> {t("sections.dates.add")}
				</Button>
			</div>
			{saveError && (
				<Alert variant="destructive" className="mb-2">
					<AlertDescription>{saveError}</AlertDescription>
				</Alert>
			)}
			<div className="space-y-2">
				{/* Birthday from DOB — read-only */}
				{person.date_of_birth && (
					<div className="flex items-center gap-3 text-sm border-bw border-line rounded-md p-2 bg-muted">
						<Badge variant="neutral">{t("sections.dates.birthday")}</Badge>
						<span className="font-mono text-[12px] text-sub flex-1">
							{person.date_of_birth}
						</span>
						<Lock className="size-3 text-sub" />
					</div>
				)}
				{data.map((d) =>
					editingId === d.id ? (
						<DateForm
							key={d.id}
							date={d}
							onSave={handleSave}
							onCancel={() => setEditingId(null)}
						/>
					) : (
						<div
							key={d.id}
							className="flex items-center gap-3 text-sm border-bw border-line rounded-md p-2"
						>
							<Badge variant="neutral">
								{t(`sections.dates.kind.${d.kind}`, { defaultValue: d.kind })}
							</Badge>
							{d.label && <span className="text-sub">{d.label}</span>}
							<span className="font-mono text-[12px] text-sub flex-1">
								{d.date_value}
							</span>
							{d.recurring && <span className="text-sub text-xs">↻</span>}
							<button
								type="button"
								onClick={() => setEditingId(d.id)}
								className="text-foreground/40 hover:text-main"
							>
								<Pencil className="size-3" />
							</button>
							<button
								type="button"
								onClick={() => setConfirmDeleteId(d.id)}
								className="text-foreground/40 hover:text-destructive"
							>
								<Trash2 className="size-3" />
							</button>
						</div>
					),
				)}
				{editingId === "new" && (
					<DateForm
						date={{}}
						onSave={handleSave}
						onCancel={() => setEditingId(null)}
					/>
				)}
				{data.length === 0 && !person.date_of_birth && editingId !== "new" && (
					<p className="text-sm text-sub">{t("sections.dates.empty")}</p>
				)}
			</div>
			<Dialog
				open={confirmDeleteId !== null}
				onOpenChange={(v) => !v && setConfirmDeleteId(null)}
			>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>{t("sections.dates.removeTitle")}</DialogTitle>
					</DialogHeader>
					{(() => {
						const d = data.find((d) => d.id === confirmDeleteId);
						return d ? (
							<p className="text-[13px] text-sub">
								<SafeTrans
									t={t}
									i18nKey={
										d.label
											? "sections.dates.removeBodyWithLabel"
											: "sections.dates.removeBody"
									}
									values={{
										kind: t(`sections.dates.kind.${d.kind}`, {
											defaultValue: d.kind,
										}),
										label: d.label,
										value: d.date_value,
									}}
									components={{ b: <span className="font-medium" /> }}
								/>
							</p>
						) : null;
					})()}
					<DialogFooter>
						<Button variant="neutral" onClick={() => setConfirmDeleteId(null)}>
							{t("sections.dates.cancel")}
						</Button>
						<Button
							variant="destructive"
							disabled={saveMutation.isPending}
							onClick={() => {
								if (confirmDeleteId !== null) {
									handleDelete(confirmDeleteId);
									setConfirmDeleteId(null);
								}
							}}
						>
							{saveMutation.isPending
								? t("sections.dates.removing")
								: t("sections.dates.remove")}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</div>
	);
}

interface ImportantDatesSectionProps {
	personId: number;
	person: Person;
}

export function ImportantDatesSection({
	personId,
	person,
}: ImportantDatesSectionProps) {
	return (
		<QueryBoundary>
			<ImportantDatesSectionInner personId={personId} person={person} />
		</QueryBoundary>
	);
}
