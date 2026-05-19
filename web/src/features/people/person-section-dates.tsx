import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Check, Lock, Pencil, Plus, Trash2, X } from "lucide-react";
import { useState } from "react";
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
import { listDatesByPerson, replaceDatesForPerson } from "#/endpoints/dates";
import { keys } from "#/query-keys";
import type { ImportantDate } from "#/schemas/date";
import type { Person } from "#/schemas/person";

function SectionHeading({ children }: { children: React.ReactNode }) {
	return (
		<h2 className="text-[11px] font-semibold uppercase tracking-widest text-zinc-400 mb-2">
			{children}
		</h2>
	);
}

const DATE_KINDS = ["Anniversary", "Memorial", "Custom"];

interface DateFormProps {
	date: Partial<ImportantDate>;
	onSave: (d: Partial<ImportantDate>) => void;
	onCancel: () => void;
}

function DateForm({ date, onSave, onCancel }: DateFormProps) {
	const [kind, setKind] = useState(date.kind ?? "Custom");
	const [label, setLabel] = useState(date.label ?? "");
	const [dateValue, setDateValue] = useState(date.date_value ?? "");
	const [recurring, setRecurring] = useState(date.recurring ?? false);

	return (
		<div className="border border-zinc-300 rounded-md p-3 bg-zinc-50 space-y-2">
			<div className="flex gap-2">
				<select
					className="h-8 border border-zinc-200 rounded-md bg-white px-2 text-sm"
					value={kind}
					onChange={(e) => setKind(e.target.value)}
				>
					{DATE_KINDS.map((k) => (
						<option key={k} value={k}>
							{k}
						</option>
					))}
				</select>
				<Input
					className="h-8 flex-1"
					value={label}
					onChange={(e) => setLabel(e.target.value)}
					placeholder="Label (optional)"
				/>
			</div>
			<div className="flex gap-2 items-center">
				<Input
					className="h-8 flex-1"
					value={dateValue}
					onChange={(e) => setDateValue(e.target.value)}
					placeholder="Date (YYYY-MM-DD or --MM-DD)"
				/>
				<label className="flex items-center gap-1 text-sm text-zinc-600 cursor-pointer">
					<input
						type="checkbox"
						checked={recurring}
						onChange={(e) => setRecurring(e.target.checked)}
					/>
					Recurring
				</label>
			</div>
			<div className="flex gap-2 justify-end">
				<Button variant="neutral" size="sm" onClick={onCancel}>
					<X className="size-3" /> Cancel
				</Button>
				<Button
					size="sm"
					disabled={!dateValue}
					onClick={() =>
						onSave({ ...date, kind, label, date_value: dateValue, recurring })
					}
				>
					<Check className="size-3" /> Save
				</Button>
			</div>
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
	const qc = useQueryClient();
	const [editingId, setEditingId] = useState<number | "new" | null>(null);
	const [saveError, setSaveError] = useState<string | null>(null);
	const [confirmDeleteId, setConfirmDeleteId] = useState<number | null>(null);

	const { data = [] } = useQuery({
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
			setSaveError(e instanceof Error ? e.message : "Failed to save"),
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
					kind: updated.kind!,
					label: updated.label ?? "",
					date_value: updated.date_value!,
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
				<SectionHeading>Important Dates</SectionHeading>
				<Button variant="neutral" size="sm" onClick={() => setEditingId("new")}>
					<Plus className="size-3" /> Add
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
					<div className="flex items-center gap-3 text-sm border border-zinc-200 rounded-md p-2 bg-zinc-50">
						<Badge variant="neutral">Birthday</Badge>
						<span className="font-mono text-[12px] text-zinc-500 flex-1">
							{person.date_of_birth}
						</span>
						<Lock className="size-3 text-zinc-300" />
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
							className="flex items-center gap-3 text-sm border border-zinc-200 rounded-md p-2"
						>
							<Badge variant="neutral">{d.kind}</Badge>
							{d.label && <span className="text-zinc-500">{d.label}</span>}
							<span className="font-mono text-[12px] text-zinc-500 flex-1">
								{d.date_value}
							</span>
							{d.recurring && <span className="text-zinc-400 text-xs">↻</span>}
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
					<p className="text-sm text-zinc-400">No important dates.</p>
				)}
			</div>
			<Dialog
				open={confirmDeleteId !== null}
				onOpenChange={(v) => !v && setConfirmDeleteId(null)}
			>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Remove date?</DialogTitle>
					</DialogHeader>
					{(() => {
						const d = data.find((d) => d.id === confirmDeleteId);
						return d ? (
							<p className="text-[13px] text-zinc-600">
								Remove the <span className="font-medium">{d.kind}</span>
								{d.label ? ` (${d.label})` : ""} date{" "}
								<span className="font-medium">{d.date_value}</span>?
							</p>
						) : null;
					})()}
					<DialogFooter>
						<Button variant="neutral" onClick={() => setConfirmDeleteId(null)}>
							Cancel
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
							{saveMutation.isPending ? "Removing…" : "Remove"}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</div>
	);
}
