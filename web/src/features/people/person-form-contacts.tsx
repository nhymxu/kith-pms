import { Plus, Trash2 } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Button } from "#/components/ui/button";
import { Input } from "#/components/ui/input";
import { Label } from "#/components/ui/label";
import type { ContactInfo } from "#/schemas/person";

type ContactRow = Omit<ContactInfo, "id" | "person_id">;

interface Props {
	value: ContactRow[];
	onChange: (rows: ContactRow[]) => void;
}

const CONTACT_TYPES = ["phone", "email", "address", "social", "other"] as const;

export function PersonFormContacts({ value, onChange }: Props) {
	const { t } = useTranslation("people");
	function add() {
		onChange([
			...value,
			{ type: "phone", value: "", label: "", position: value.length },
		]);
	}

	function remove(i: number) {
		onChange(value.filter((_, idx) => idx !== i));
	}

	function update(i: number, field: keyof ContactRow, v: string | number) {
		onChange(
			value.map((row, idx) => (idx === i ? { ...row, [field]: v } : row)),
		);
	}

	return (
		<div className="space-y-3">
			<div className="flex items-center justify-between">
				<Label className="text-sm font-medium">
					{t("form.contacts.title")}
				</Label>
				<Button type="button" variant="neutral" size="sm" onClick={add}>
					<Plus className="size-3" /> {t("form.contacts.add")}
				</Button>
			</div>
			{value.length === 0 && (
				<p className="text-xs font-base text-sub">{t("form.contacts.empty")}</p>
			)}
			{value.map((row, i) => (
				<div
					// biome-ignore lint/suspicious/noArrayIndexKey: no stable id on unsaved rows
					key={i}
					className="grid grid-cols-[100px_1fr_1fr_32px] gap-2 items-end"
				>
					<div>
						<Label className="text-xs">{t("form.contacts.type")}</Label>
						<select
							className="w-full h-10 border-field-bw border-field-line rounded-md bg-field px-2 text-sm font-base"
							value={row.type}
							onChange={(e) => update(i, "type", e.target.value)}
						>
							{CONTACT_TYPES.map((ct) => (
								<option key={ct} value={ct}>
									{t(`types.contact.${ct}`)}
								</option>
							))}
						</select>
					</div>
					<div>
						<Label className="text-xs">{t("form.contacts.value")}</Label>
						<Input
							value={row.value}
							onChange={(e) => update(i, "value", e.target.value)}
							placeholder={t("form.contacts.valuePlaceholder")}
						/>
					</div>
					<div>
						<Label className="text-xs">{t("form.contacts.label")}</Label>
						<Input
							value={row.label}
							onChange={(e) => update(i, "label", e.target.value)}
							placeholder={t("form.contacts.labelPlaceholder")}
						/>
					</div>
					<Button
						type="button"
						variant="destructive"
						size="icon"
						className="self-end"
						onClick={() => remove(i)}
					>
						<Trash2 className="size-3" />
					</Button>
				</div>
			))}
		</div>
	);
}
