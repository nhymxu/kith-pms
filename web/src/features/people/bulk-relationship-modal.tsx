import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { X } from "lucide-react";
import { useState } from "react";
import { Button } from "#/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import { Input } from "#/components/ui/input";
import { listPeople } from "#/endpoints/people";
import { listRelationshipTypes } from "#/endpoints/relationship-types";
import { bulkCreateRelationships } from "#/endpoints/relationships";
import { formatPersonName } from "#/lib/format-person-name";
import { keys } from "#/query-keys";

interface PendingPair {
	toPersonId: number;
	toPersonName: string;
	typeId: number;
	typeName: string;
}

interface BulkRelationshipModalProps {
	fromPersonId: number;
	open: boolean;
	onClose: () => void;
	onSuccess: () => void;
}

export function BulkRelationshipModal({
	fromPersonId,
	open,
	onClose,
	onSuccess,
}: BulkRelationshipModalProps) {
	const qc = useQueryClient();
	const [pending, setPending] = useState<PendingPair[]>([]);
	const [selectedTypeId, setSelectedTypeId] = useState<string>("");
	const [personSearch, setPersonSearch] = useState("");
	const [selectedPeople, setSelectedPeople] = useState<
		Array<{ id: number; name: string }>
	>([]);
	const [resultMsg, setResultMsg] = useState<string | null>(null);

	const { data: types } = useQuery({
		queryKey: keys.relationshipTypes.list(),
		queryFn: listRelationshipTypes,
		enabled: open,
	});

	const { data: searchResults } = useQuery({
		queryKey: keys.people.list({ q: personSearch || undefined }),
		queryFn: () => listPeople({ q: personSearch || undefined, page_size: 10 }),
		enabled: personSearch.length > 0,
	});

	const mutation = useMutation({
		mutationFn: () =>
			bulkCreateRelationships(
				fromPersonId,
				pending.map((p) => ({
					to_person_id: p.toPersonId,
					relationship_type_id: p.typeId,
				})),
			),
		onSuccess: (data) => {
			setResultMsg(`Created ${data.created} relationships`);
			qc.invalidateQueries({
				queryKey: keys.relationships.byPerson(fromPersonId),
			});
			onSuccess();
			handleClose();
		},
		onError: () => setResultMsg("Failed to save relationships"),
	});

	function handleAddToList() {
		if (!selectedTypeId || selectedPeople.length === 0) return;

		const typeIdNum = Number(selectedTypeId);
		const typeName =
			types?.find((t) => t.id === typeIdNum)?.name ?? String(selectedTypeId);

		const newPairs = selectedPeople
			.filter(
				(p) =>
					!pending.some(
						(existing) =>
							existing.toPersonId === p.id && existing.typeId === typeIdNum,
					),
			)
			.map((p) => ({
				toPersonId: p.id,
				toPersonName: p.name,
				typeId: typeIdNum,
				typeName,
			}));

		setPending((prev) => [...prev, ...newPairs]);
		setSelectedPeople([]);
		setPersonSearch("");
	}

	function togglePerson(person: { id: number; name: string }) {
		setSelectedPeople((prev) =>
			prev.some((p) => p.id === person.id)
				? prev.filter((p) => p.id !== person.id)
				: [...prev, person],
		);
	}

	function removePair(idx: number) {
		setPending((prev) => prev.filter((_, i) => i !== idx));
	}

	function handleClose() {
		setPending([]);
		setSelectedTypeId("");
		setPersonSearch("");
		setSelectedPeople([]);
		setResultMsg(null);
		onClose();
	}

	const filteredResults = searchResults?.items.filter(
		(p) => p.id !== fromPersonId,
	);

	return (
		<Dialog open={open} onOpenChange={(v) => !v && handleClose()}>
			<DialogContent className="max-w-lg">
				<DialogHeader>
					<DialogTitle>Add multiple relationships</DialogTitle>
				</DialogHeader>

				<div className="space-y-4">
					<div className="flex gap-2 items-end">
						<div className="flex-1 space-y-1.5">
							<p className="text-xs font-medium text-zinc-600">Type</p>
							<select
								value={selectedTypeId}
								onChange={(e) => setSelectedTypeId(e.target.value)}
								className="w-full h-9 border border-zinc-200 rounded-md bg-white px-2 text-sm"
							>
								<option value="">Select type…</option>
								{types?.map((t) => (
									<option key={t.id} value={t.id}>
										{t.name}
									</option>
								))}
							</select>
						</div>
						<Button
							size="sm"
							variant="neutral"
							disabled={!selectedTypeId || selectedPeople.length === 0}
							onClick={handleAddToList}
						>
							Add to list
						</Button>
					</div>

					<div className="space-y-1.5">
						<p className="text-xs font-medium text-zinc-600">
							People (select multiple)
						</p>
						<Input
							placeholder="Search by name…"
							value={personSearch}
							onChange={(e) => setPersonSearch(e.target.value)}
						/>
						{selectedPeople.length > 0 && (
							<div className="flex flex-wrap gap-1">
								{selectedPeople.map((p) => (
									<span
										key={p.id}
										className="inline-flex items-center gap-1 text-xs border border-zinc-300 rounded px-1.5 py-0.5 bg-zinc-50"
									>
										{p.name}
										<button
											type="button"
											onClick={() => togglePerson(p)}
											className="text-zinc-400 hover:text-zinc-700"
										>
											<X className="size-3" />
										</button>
									</span>
								))}
							</div>
						)}
						{filteredResults && filteredResults.length > 0 && (
							<div className="border border-zinc-200 rounded-md bg-white divide-y divide-zinc-100 max-h-36 overflow-y-auto">
								{filteredResults.map((p) => {
									const isSelected = selectedPeople.some(
										(sel) => sel.id === p.id,
									);
									return (
										<button
											key={p.id}
											type="button"
											onClick={() =>
												togglePerson({
													id: p.id,
													name: formatPersonName(p.name, p.nickname),
												})
											}
											className={`w-full text-left px-3 py-2 text-sm hover:bg-zinc-50 ${isSelected ? "bg-zinc-100 font-medium" : ""}`}
										>
											{formatPersonName(p.name, p.nickname)}
										</button>
									);
								})}
							</div>
						)}
					</div>

					{pending.length > 0 && (
						<div className="border border-zinc-200 rounded-md divide-y divide-zinc-100 max-h-40 overflow-y-auto">
							{pending.map((pair, idx) => (
								<div
									key={`${pair.typeId}-${pair.toPersonId}`}
									className="flex items-center gap-2 px-3 py-2 text-sm"
								>
									<span className="text-xs border border-zinc-300 rounded px-1.5 py-0.5 bg-zinc-50 shrink-0">
										{pair.typeName}
									</span>
									<span className="flex-1 text-zinc-700">
										{pair.toPersonName}
									</span>
									<button
										type="button"
										onClick={() => removePair(idx)}
										className="text-zinc-400 hover:text-zinc-700"
									>
										<X className="size-3" />
									</button>
								</div>
							))}
						</div>
					)}

					{resultMsg && (
						<p className="text-sm text-zinc-700 font-medium">{resultMsg}</p>
					)}
				</div>

				<DialogFooter>
					<Button variant="neutral" onClick={handleClose}>
						Cancel
					</Button>
					<Button
						disabled={pending.length === 0 || mutation.isPending}
						onClick={() => mutation.mutate()}
					>
						{mutation.isPending
							? "Saving…"
							: `Save ${pending.length} relationship${pending.length !== 1 ? "s" : ""}`}
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
