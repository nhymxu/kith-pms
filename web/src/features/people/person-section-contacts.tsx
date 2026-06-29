import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Check, Pencil, Plus, Trash2, X } from "lucide-react";
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
import { updatePerson } from "#/endpoints/people";
import { keys } from "#/query-keys";
import type { ContactInfo, Person } from "#/schemas/person";

function SectionHeading({ children }: { children: React.ReactNode }) {
	return (
		<h2 className="text-[11px] font-semibold uppercase tracking-widest text-zinc-400 mb-2">
			{children}
		</h2>
	);
}

const CONTACT_TYPES = ["phone", "email", "address", "social", "other"];

interface EditRowProps {
	contact: Partial<ContactInfo>;
	onSave: (c: Partial<ContactInfo>) => void;
	onCancel: () => void;
}

function ContactEditRow({ contact, onSave, onCancel }: EditRowProps) {
	const [type, setType] = useState(contact.type ?? "phone");
	const [value, setValue] = useState(contact.value ?? "");
	const [label, setLabel] = useState(contact.label ?? "");

	return (
		<div className="flex flex-wrap gap-2 items-center border border-zinc-300 rounded-md p-2 bg-zinc-50">
			<select
				className="h-8 border border-zinc-200 rounded-md bg-white px-2 text-sm"
				value={type}
				onChange={(e) => setType(e.target.value)}
			>
				{CONTACT_TYPES.map((t) => (
					<option key={t} value={t}>
						{t}
					</option>
				))}
			</select>
			<Input
				className="h-8 flex-1 min-w-[120px]"
				value={value}
				onChange={(e) => setValue(e.target.value)}
				placeholder="Value"
			/>
			<Input
				className="h-8 w-24"
				value={label}
				onChange={(e) => setLabel(e.target.value)}
				placeholder="Label"
			/>
			<button
				type="button"
				onClick={() => onSave({ ...contact, type, value, label })}
				className="text-green-600 hover:text-green-700"
			>
				<Check className="size-4" />
			</button>
			<button
				type="button"
				onClick={onCancel}
				className="text-zinc-400 hover:text-zinc-600"
			>
				<X className="size-4" />
			</button>
		</div>
	);
}

interface ContactsSectionProps {
	person: Person;
}

export function ContactsSection({ person }: ContactsSectionProps) {
	const qc = useQueryClient();
	const [editingId, setEditingId] = useState<number | "new" | null>(null);
	const [confirmContactId, setConfirmContactId] = useState<number | null>(null);
	const [saveError, setSaveError] = useState<string | null>(null);

	const saveMutation = useMutation({
		mutationFn: (contacts: ContactInfo[]) =>
			updatePerson(person.id, {
				name: person.name,
				nickname: person.nickname,
				gender: person.gender ?? "",
				relationship_type: person.relationship_type,
				date_of_birth: person.date_of_birth ?? "",
				create_birthday_reminder: person.has_birthday_reminder,
				last_contact_at: person.last_contact_at ?? null,
				other_notes: person.other_notes,
				contacts: contacts.map((c, i) => ({
					type: c.type,
					value: c.value,
					label: c.label,
					position: i,
				})),
				locations: person.locations.map((l, i) => ({
					type: l.type,
					address: l.address,
					city: l.city,
					country: l.country,
					postal_code: l.postal_code,
					position: i,
				})),
			}),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.people.detail(person.id) });
			setEditingId(null);
			setSaveError(null);
		},
		onError: (e) =>
			setSaveError(e instanceof Error ? e.message : "Failed to save"),
	});

	function handleSave(updated: Partial<ContactInfo>) {
		let contacts: ContactInfo[];
		if (editingId === "new") {
			contacts = [
				...person.contacts,
				{
					id: 0,
					person_id: person.id,
					type: updated.type ?? "",
					value: updated.value ?? "",
					label: updated.label ?? "",
					position: person.contacts.length,
				},
			];
		} else {
			contacts = person.contacts.map((c) =>
				c.id === editingId ? { ...c, ...updated } : c,
			);
		}
		saveMutation.mutate(contacts);
	}

	function handleDelete(id: number) {
		const contacts = person.contacts.filter((c) => c.id !== id);
		saveMutation.mutate(contacts, {
			onSuccess: () => setConfirmContactId(null),
		});
	}

	return (
		<div>
			<div className="flex items-center justify-between mb-2">
				<SectionHeading>Contacts</SectionHeading>
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
				{person.contacts.map((c) =>
					editingId === c.id ? (
						<ContactEditRow
							key={c.id}
							contact={c}
							onSave={handleSave}
							onCancel={() => setEditingId(null)}
						/>
					) : (
						<div
							key={c.id}
							className="flex gap-3 text-sm border border-zinc-200 rounded-md p-2 items-center"
						>
							<Badge variant="neutral">{c.type}</Badge>
							<span className="font-base flex-1">{c.value}</span>
							{c.label && <span className="text-zinc-400">{c.label}</span>}
							<button
								type="button"
								onClick={() => setEditingId(c.id)}
								className="text-foreground/40 hover:text-main"
							>
								<Pencil className="size-3" />
							</button>
							<button
								type="button"
								onClick={() => setConfirmContactId(c.id)}
								className="text-foreground/40 hover:text-destructive"
							>
								<Trash2 className="size-3" />
							</button>
						</div>
					),
				)}
				{editingId === "new" && (
					<ContactEditRow
						contact={{}}
						onSave={handleSave}
						onCancel={() => setEditingId(null)}
					/>
				)}
				{person.contacts.length === 0 && editingId !== "new" && (
					<p className="text-sm text-zinc-400">No contacts.</p>
				)}
			</div>
			<Dialog
				open={confirmContactId !== null}
				onOpenChange={(v) => !v && setConfirmContactId(null)}
			>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Remove contact?</DialogTitle>
					</DialogHeader>
					{(() => {
						const c = person.contacts.find((c) => c.id === confirmContactId);
						return c ? (
							<p className="text-[13px] text-zinc-600">
								Remove the <span className="font-medium">{c.type}</span> contact{" "}
								<span className="font-medium">{c.value}</span>
								{c.label ? ` (${c.label})` : ""}?
							</p>
						) : null;
					})()}
					<DialogFooter>
						<Button variant="neutral" onClick={() => setConfirmContactId(null)}>
							Cancel
						</Button>
						<Button
							variant="destructive"
							disabled={saveMutation.isPending}
							onClick={() =>
								confirmContactId !== null && handleDelete(confirmContactId)
							}
						>
							{saveMutation.isPending ? "Removing…" : "Remove"}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</div>
	);
}
