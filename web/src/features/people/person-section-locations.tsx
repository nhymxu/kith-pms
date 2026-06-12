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
import type { Location, Person } from "#/schemas/person";

function SectionHeading({ children }: { children: React.ReactNode }) {
	return (
		<h2 className="text-[11px] font-semibold uppercase tracking-widest text-zinc-400 mb-2">
			{children}
		</h2>
	);
}

const LOCATION_TYPES = ["home", "work", "other"];

interface EditRowProps {
	location: Partial<Location>;
	onSave: (l: Partial<Location>) => void;
	onCancel: () => void;
}

function LocationEditRow({ location, onSave, onCancel }: EditRowProps) {
	const [type, setType] = useState(location.type ?? "home");
	const [address, setAddress] = useState(location.address ?? "");
	const [city, setCity] = useState(location.city ?? "");
	const [country, setCountry] = useState(location.country ?? "");
	const [postalCode, setPostalCode] = useState(location.postal_code ?? "");

	return (
		<div className="border border-zinc-300 rounded-md p-3 bg-zinc-50 space-y-2">
			<div className="flex gap-2 items-center">
				<select
					className="h-8 border border-zinc-200 rounded-md bg-white px-2 text-sm"
					value={type}
					onChange={(e) => setType(e.target.value)}
				>
					{LOCATION_TYPES.map((t) => (
						<option key={t} value={t}>
							{t}
						</option>
					))}
				</select>
				<div className="flex gap-1 ml-auto">
					<button
						type="button"
						onClick={() =>
							onSave({
								...location,
								type,
								address,
								city,
								country,
								postal_code: postalCode,
							})
						}
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
			</div>
			<Input
				className="h-8"
				value={address}
				onChange={(e) => setAddress(e.target.value)}
				placeholder="Address"
			/>
			<div className="flex gap-2">
				<Input
					className="h-8 flex-1"
					value={city}
					onChange={(e) => setCity(e.target.value)}
					placeholder="City"
				/>
				<Input
					className="h-8 w-24"
					value={postalCode}
					onChange={(e) => setPostalCode(e.target.value)}
					placeholder="Postal code"
				/>
				<Input
					className="h-8 flex-1"
					value={country}
					onChange={(e) => setCountry(e.target.value)}
					placeholder="Country"
				/>
			</div>
		</div>
	);
}

interface LocationsSectionProps {
	person: Person;
}

export function LocationsSection({ person }: LocationsSectionProps) {
	const qc = useQueryClient();
	const [editingId, setEditingId] = useState<number | "new" | null>(null);
	const [saveError, setSaveError] = useState<string | null>(null);
	const [confirmDeleteId, setConfirmDeleteId] = useState<number | null>(null);

	const saveMutation = useMutation({
		mutationFn: (locations: Location[]) =>
			updatePerson(person.id, {
				name: person.name,
				nickname: person.nickname,
				gender: person.gender ?? "",
				relationship_type: person.relationship_type,
				date_of_birth: person.date_of_birth ?? "",
				last_contact_at: person.last_contact_at ?? null,
				other_notes: person.other_notes,
				contacts: person.contacts.map((c, i) => ({
					type: c.type,
					value: c.value,
					label: c.label,
					position: i,
				})),
				locations: locations.map((l, i) => ({
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

	function handleSave(updated: Partial<Location>) {
		let locations: Location[];
		if (editingId === "new") {
			locations = [
				...person.locations,
				{
					id: 0,
					person_id: person.id,
					type: updated.type ?? "",
					address: updated.address ?? "",
					city: updated.city ?? "",
					country: updated.country ?? "",
					postal_code: updated.postal_code ?? "",
					position: person.locations.length,
				},
			];
		} else {
			locations = person.locations.map((l) =>
				l.id === editingId ? { ...l, ...updated } : l,
			);
		}
		saveMutation.mutate(locations);
	}

	function handleDelete(id: number) {
		saveMutation.mutate(person.locations.filter((l) => l.id !== id));
	}

	return (
		<div>
			<div className="flex items-center justify-between mb-2">
				<SectionHeading>Locations</SectionHeading>
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
				{person.locations.map((l) =>
					editingId === l.id ? (
						<LocationEditRow
							key={l.id}
							location={l}
							onSave={handleSave}
							onCancel={() => setEditingId(null)}
						/>
					) : (
						<div
							key={l.id}
							className="text-sm border border-zinc-200 rounded-md p-3 space-y-1"
						>
							<div className="flex items-center gap-2">
								<Badge variant="neutral">{l.type}</Badge>
								<div className="flex gap-1 ml-auto">
									<button
										type="button"
										onClick={() => setEditingId(l.id)}
										className="text-foreground/40 hover:text-main"
									>
										<Pencil className="size-3" />
									</button>
									<button
										type="button"
										onClick={() => setConfirmDeleteId(l.id)}
										className="text-foreground/40 hover:text-destructive"
									>
										<Trash2 className="size-3" />
									</button>
								</div>
							</div>
							<p className="font-base">
								{[l.address, l.city, l.country, l.postal_code]
									.filter(Boolean)
									.join(", ")}
							</p>
						</div>
					),
				)}
				{editingId === "new" && (
					<LocationEditRow
						location={{}}
						onSave={handleSave}
						onCancel={() => setEditingId(null)}
					/>
				)}
				{person.locations.length === 0 && editingId !== "new" && (
					<p className="text-sm text-zinc-400">No locations.</p>
				)}
			</div>
			<Dialog
				open={confirmDeleteId !== null}
				onOpenChange={(v) => !v && setConfirmDeleteId(null)}
			>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Remove location?</DialogTitle>
					</DialogHeader>
					{(() => {
						const l = person.locations.find((l) => l.id === confirmDeleteId);
						return l ? (
							<p className="text-[13px] text-zinc-600">
								Remove the <span className="font-medium">{l.type}</span>{" "}
								location{l.address ? ` at ${l.address}` : ""}?
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
