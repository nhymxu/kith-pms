import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { Link2, Search, Trash2, X } from "lucide-react";
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
import { Label } from "#/components/ui/label";
import {
	attachRelationship,
	detachRelationship,
	listPeople,
	listRelationships,
} from "#/endpoints/people";
import { listRelationshipTypes } from "#/endpoints/relationship-types";
import { keys } from "#/query-keys";

function SectionHeading({ children }: { children: React.ReactNode }) {
	return (
		<h2 className="text-[11px] font-semibold uppercase tracking-widest text-zinc-400 mb-2">
			{children}
		</h2>
	);
}

interface RelationshipsSectionProps {
	personId: number;
}

export function RelationshipsSection({ personId }: RelationshipsSectionProps) {
	const qc = useQueryClient();
	const [addOpen, setAddOpen] = useState(false);
	const [typeId, setTypeId] = useState<number | "">("");
	const [otherPersonId, setOtherPersonId] = useState<number | null>(null);
	const [otherPersonName, setOtherPersonName] = useState("");
	const [personSearch, setPersonSearch] = useState("");
	const [notes, setNotes] = useState("");
	const [err, setErr] = useState<string | null>(null);
	const [confirmRelId, setConfirmRelId] = useState<number | null>(null);

	const { data: rels } = useQuery({
		queryKey: keys.people.relationships(personId),
		queryFn: () => listRelationships(personId),
	});
	const { data: types } = useQuery({
		queryKey: keys.relationshipTypes.list(),
		queryFn: listRelationshipTypes,
	});
	const { data: personResults } = useQuery({
		queryKey: keys.people.list({ q: personSearch || undefined }),
		queryFn: () => listPeople({ q: personSearch || undefined, page_size: 10 }),
		enabled: personSearch.length > 0,
	});

	const attach = useMutation({
		mutationFn: () =>
			attachRelationship(personId, {
				relationship_type_id: Number(typeId),
				to_person_id: otherPersonId ?? 0,
				notes,
			}),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.people.relationships(personId) });
			setAddOpen(false);
			setTypeId("");
			setOtherPersonId(null);
			setOtherPersonName("");
			setPersonSearch("");
			setNotes("");
		},
		onError: (e) => setErr(e instanceof Error ? e.message : "Failed"),
	});

	const detach = useMutation({
		mutationFn: (relId: number) => detachRelationship(personId, relId),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.people.relationships(personId) });
			setConfirmRelId(null);
		},
	});

	const confirmRel = rels?.find((r) => r.id === confirmRelId);

	return (
		<div>
			<div className="flex items-center justify-between mb-2">
				<SectionHeading>Relationships</SectionHeading>
				<Button variant="neutral" size="sm" onClick={() => setAddOpen(true)}>
					<Link2 className="size-3" /> Add
				</Button>
			</div>
			{!rels?.length ? (
				<p className="text-sm text-zinc-400">No relationships yet.</p>
			) : (
				<div className="space-y-2">
					{rels.map((r) => (
						<div
							key={r.id}
							className="flex items-center gap-3 border border-zinc-200 rounded-md p-2 text-sm"
						>
							<Badge variant="neutral">{r.type_name}</Badge>
							<Link
								to="/people/$personId"
								params={{ personId: String(r.other_person_id) }}
								className="font-medium hover:underline flex-1"
							>
								{r.other_person_name}
							</Link>
							{r.notes && (
								<span className="text-zinc-400 text-xs truncate max-w-[140px]">
									{r.notes}
								</span>
							)}
							<button
								type="button"
								onClick={() => setConfirmRelId(r.id)}
								className="text-foreground/40 hover:text-destructive"
							>
								<Trash2 className="size-3" />
							</button>
						</div>
					))}
				</div>
			)}

			<Dialog open={addOpen} onOpenChange={(v) => !v && setAddOpen(false)}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Add relationship</DialogTitle>
					</DialogHeader>
					{err && (
						<Alert variant="destructive">
							<AlertDescription>{err}</AlertDescription>
						</Alert>
					)}
					<div className="space-y-3">
						<div>
							<Label>Type</Label>
							<select
								className="w-full h-10 border border-zinc-200 rounded-md bg-white px-2 text-sm"
								value={typeId}
								onChange={(e) => setTypeId(Number(e.target.value))}
							>
								<option value="">Select…</option>
								{types?.map((t) => (
									<option key={t.id} value={t.id}>
										{t.name}
									</option>
								))}
							</select>
						</div>
						<div className="space-y-1">
							<Label>Person</Label>
							{otherPersonId ? (
								<div className="flex items-center gap-2 border border-zinc-200 rounded-md px-3 py-2 text-sm">
									<span className="flex-1 font-medium">{otherPersonName}</span>
									<button
										type="button"
										onClick={() => {
											setOtherPersonId(null);
											setOtherPersonName("");
											setPersonSearch("");
										}}
										className="text-zinc-400 hover:text-destructive"
									>
										<X className="size-3" />
									</button>
								</div>
							) : (
								<div className="space-y-1">
									<div className="relative">
										<Search className="absolute left-2.5 top-2.5 size-3.5 text-zinc-400" />
										<input
											type="text"
											value={personSearch}
											onChange={(e) => setPersonSearch(e.target.value)}
											placeholder="Search by name…"
											className="h-9 w-full border border-zinc-200 rounded-md bg-white pl-8 pr-3 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-600"
										/>
									</div>
									{personResults?.items && personResults.items.length > 0 && (
										<div className="border border-zinc-200 rounded-md bg-white divide-y divide-zinc-100 max-h-40 overflow-y-auto">
											{personResults.items
												.filter((p) => p.id !== personId)
												.map((p) => (
													<button
														key={p.id}
														type="button"
														onClick={() => {
															setOtherPersonId(p.id);
															setOtherPersonName(p.name);
															setPersonSearch("");
														}}
														className="w-full text-left px-3 py-2 text-sm hover:bg-zinc-50"
													>
														{p.name}
													</button>
												))}
										</div>
									)}
								</div>
							)}
						</div>
						<div>
							<Label>Notes</Label>
							<Input
								value={notes}
								onChange={(e) => setNotes(e.target.value)}
								placeholder="Optional notes"
							/>
						</div>
					</div>
					<DialogFooter>
						<Button variant="neutral" onClick={() => setAddOpen(false)}>
							Cancel
						</Button>
						<Button
							disabled={attach.isPending || !typeId || !otherPersonId}
							onClick={() => attach.mutate()}
						>
							Save
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>

			<Dialog
				open={confirmRelId !== null}
				onOpenChange={(v) => !v && setConfirmRelId(null)}
			>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Remove relationship?</DialogTitle>
					</DialogHeader>
					{confirmRel && (
						<p className="text-[13px] text-zinc-600">
							Remove the{" "}
							<span className="font-medium">{confirmRel.type_name}</span>{" "}
							relationship with{" "}
							<span className="font-medium">
								{confirmRel.other_person_name}
							</span>
							?
						</p>
					)}
					<DialogFooter>
						<Button variant="neutral" onClick={() => setConfirmRelId(null)}>
							Cancel
						</Button>
						<Button
							variant="destructive"
							disabled={detach.isPending}
							onClick={() =>
								confirmRelId !== null && detach.mutate(confirmRelId)
							}
						>
							{detach.isPending ? "Removing…" : "Remove"}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</div>
	);
}
