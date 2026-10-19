import { useQuery } from "@tanstack/react-query";
import { X } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Input } from "#/components/ui/input";
import { listPeople } from "#/endpoints/people";
import { formatPersonName } from "#/lib/format-person-name";
import { keys } from "#/query-keys";

interface PersonSinglePickerProps {
	value: { id: number; name: string } | null;
	onChange: (person: { id: number; name: string } | null) => void;
}

export function PersonSinglePicker({
	value,
	onChange,
}: PersonSinglePickerProps) {
	const { t } = useTranslation("people");
	const [search, setSearch] = useState("");

	const { data: results } = useQuery({
		queryKey: keys.people.list({ q: search || undefined }),
		queryFn: () => listPeople({ q: search || undefined, page_size: 10 }),
		enabled: search.length > 0,
	});

	function handleSelect(person: { id: number; name: string }) {
		onChange(person);
		setSearch("");
	}

	function handleClear() {
		onChange(null);
		setSearch("");
	}

	return (
		<div className="space-y-1.5">
			{value ? (
				<div className="inline-flex items-center gap-1 text-sm border-bw border-chip-line rounded px-2 py-1 bg-chip">
					<span>{value.name}</span>
					<button
						type="button"
						onClick={handleClear}
						className="text-sub hover:text-ink"
					>
						<X className="size-3" />
					</button>
				</div>
			) : (
				<Input
					placeholder={t("picker.searchPlaceholder")}
					value={search}
					onChange={(e) => setSearch(e.target.value)}
				/>
			)}
			{!value && results && results.items.length > 0 && (
				<div className="border-bw border-line rounded-md bg-panel divide-y divide-line-soft max-h-36 overflow-y-auto">
					{results.items.map((p) => (
						<button
							key={p.id}
							type="button"
							onClick={() =>
								handleSelect({
									id: p.id,
									name: formatPersonName(p.name, p.nickname),
								})
							}
							className="w-full text-left px-3 py-2 text-sm hover:bg-chip"
						>
							{formatPersonName(p.name, p.nickname)}
						</button>
					))}
				</div>
			)}
		</div>
	);
}
