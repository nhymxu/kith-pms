import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Check, ExternalLink, Pencil, Plus, Trash2, X } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
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
import { updatePerson } from "#/endpoints/people";
import { apiErrorMessage } from "#/lib/api-error-message";
import { keys } from "#/query-keys";
import type { ContactInfo, Person } from "#/schemas/person";

function SectionHeading({ children }: { children: React.ReactNode }) {
	return (
		<h2 className="text-[11px] font-semibold uppercase tracking-widest text-sub mb-2">
			{children}
		</h2>
	);
}

const CONTACT_TYPES = ["phone", "email", "address", "social", "other"] as const;

const SOCIAL_BASE_URLS: Record<string, string> = {
	twitter: "https://twitter.com/",
	x: "https://x.com/",
	github: "https://github.com/",
	linkedin: "https://linkedin.com/in/",
	instagram: "https://instagram.com/",
	facebook: "https://facebook.com/",
	youtube: "https://youtube.com/@",
	tiktok: "https://tiktok.com/@",
	telegram: "https://t.me/",
	threads: "https://www.threads.com/@",
};

function resolveSocialUrl(value: string, label: string): string | null {
	if (/^https?:\/\//i.test(value)) return value;
	const base = SOCIAL_BASE_URLS[label.toLowerCase().trim()];
	if (!base) return null;
	const username = value.startsWith("@") ? value.slice(1) : value;
	return base + username;
}

interface EditRowProps {
	contact: Partial<ContactInfo>;
	onSave: (c: Partial<ContactInfo>) => void;
	onCancel: () => void;
}

function ContactEditRow({ contact, onSave, onCancel }: EditRowProps) {
	const { t } = useTranslation("people");
	const [type, setType] = useState(contact.type ?? "phone");
	const [value, setValue] = useState(contact.value ?? "");
	const [label, setLabel] = useState(contact.label ?? "");

	return (
		<div className="flex flex-wrap gap-2 items-center border-bw border-line rounded-md p-2 bg-muted">
			<select
				className="h-8 border-field-bw border-field-line rounded-md bg-field px-2 text-sm"
				value={type}
				onChange={(e) => setType(e.target.value)}
			>
				{CONTACT_TYPES.map((ct) => (
					<option key={ct} value={ct}>
						{t(`types.contact.${ct}`)}
					</option>
				))}
			</select>
			<Input
				className="h-8 flex-1 min-w-[120px]"
				value={value}
				onChange={(e) => setValue(e.target.value)}
				placeholder={t("sections.contacts.valuePlaceholder")}
			/>
			<Input
				className="h-8 w-24"
				value={label}
				onChange={(e) => setLabel(e.target.value)}
				placeholder={t("sections.contacts.labelPlaceholder")}
			/>
			<button
				type="button"
				onClick={() => onSave({ ...contact, type, value, label })}
				className="text-success-fg hover:opacity-80"
			>
				<Check className="size-4" />
			</button>
			<button
				type="button"
				onClick={onCancel}
				className="text-sub hover:text-ink"
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
	const { t } = useTranslation("people");
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
				date_of_birth: person.date_of_birth ?? "",
				create_birthday_reminder: person.has_birthday_reminder,
				last_contact_at: person.last_contact_at ?? null,
				bio: person.bio,
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
			setSaveError(apiErrorMessage(e, t("sections.contacts.saveFailed"))),
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
				<SectionHeading>{t("sections.contacts.title")}</SectionHeading>
				<Button variant="neutral" size="sm" onClick={() => setEditingId("new")}>
					<Plus className="size-3" /> {t("sections.contacts.add")}
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
							className="flex gap-3 text-sm border-bw border-line rounded-md p-2 items-center"
						>
							<Badge variant="neutral">
								{t(`types.contact.${c.type}`, { defaultValue: c.type })}
							</Badge>
							{c.type === "social" ? (
								(() => {
									const url = resolveSocialUrl(c.value, c.label ?? "");
									return url ? (
										<a
											href={url}
											target="_blank"
											rel="noopener noreferrer"
											className="font-base flex-1 flex items-center gap-1 hover:underline text-inherit"
										>
											{c.value}
											<ExternalLink className="size-3 text-sub shrink-0" />
										</a>
									) : (
										<span className="font-base flex-1">{c.value}</span>
									);
								})()
							) : (
								<span className="font-base flex-1">{c.value}</span>
							)}
							{c.label && <span className="text-sub">{c.label}</span>}
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
					<p className="text-sm text-sub">{t("sections.contacts.empty")}</p>
				)}
			</div>
			<Dialog
				open={confirmContactId !== null}
				onOpenChange={(v) => !v && setConfirmContactId(null)}
			>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>{t("sections.contacts.removeTitle")}</DialogTitle>
					</DialogHeader>
					{(() => {
						const c = person.contacts.find((c) => c.id === confirmContactId);
						return c ? (
							<p className="text-[13px] text-sub">
								<SafeTrans
									t={t}
									i18nKey={
										c.label
											? "sections.contacts.removeBodyWithLabel"
											: "sections.contacts.removeBody"
									}
									values={{
										type: t(`types.contact.${c.type}`, {
											defaultValue: c.type,
										}),
										value: c.value,
										label: c.label,
									}}
									components={{ b: <span className="font-medium" /> }}
								/>
							</p>
						) : null;
					})()}
					<DialogFooter>
						<Button variant="neutral" onClick={() => setConfirmContactId(null)}>
							{t("sections.contacts.cancel")}
						</Button>
						<Button
							variant="destructive"
							disabled={saveMutation.isPending}
							onClick={() =>
								confirmContactId !== null && handleDelete(confirmContactId)
							}
						>
							{saveMutation.isPending
								? t("sections.contacts.removing")
								: t("sections.contacts.remove")}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</div>
	);
}
