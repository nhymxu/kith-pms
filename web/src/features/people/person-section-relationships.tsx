import {
	useMutation,
	useQuery,
	useQueryClient,
	useSuspenseQuery,
} from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { Link2, ListPlus, Search, Trash2, X } from "lucide-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { QueryBoundary } from "#/components/query-boundary";
import { SafeTrans } from "#/components/safe-trans";
import { Alert, AlertDescription } from "#/components/ui/alert";
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
import { getMe } from "#/endpoints/me";
import {
	attachRelationship,
	detachRelationship,
	listPeople,
	listRelationships,
} from "#/endpoints/people";
import { listRelationshipTypes } from "#/endpoints/relationship-types";
import { apiErrorMessage } from "#/lib/api-error-message";
import { formatPersonName } from "#/lib/format-person-name";
import { keys } from "#/query-keys";
import { BulkRelationshipModal } from "./bulk-relationship-modal";
import {
	RelationshipLabel,
	useRelationshipDisplay,
} from "./relationship-label";

function SectionHeading({ children }: { children: React.ReactNode }) {
	return (
		<h2 className="text-[11px] font-semibold uppercase tracking-widest text-sub mb-2">
			{children}
		</h2>
	);
}

interface RelationshipsSectionProps {
	personId: number;
	personName: string;
}

export function RelationshipsSection({
	personId,
	personName,
}: RelationshipsSectionProps) {
	return (
		<QueryBoundary>
			<RelationshipsSectionInner personId={personId} personName={personName} />
		</QueryBoundary>
	);
}

function RelationshipsSectionInner({
	personId,
	personName,
}: RelationshipsSectionProps) {
	const { t } = useTranslation("people");
	const qc = useQueryClient();
	const display = useRelationshipDisplay();
	const [addOpen, setAddOpen] = useState(false);
	const [bulkOpen, setBulkOpen] = useState(false);
	const [typeId, setTypeId] = useState<number | "">("");
	const [otherPersonId, setOtherPersonId] = useState<number | null>(null);
	const [otherPersonName, setOtherPersonName] = useState("");
	const [personSearch, setPersonSearch] = useState("");
	const [notes, setNotes] = useState("");
	const [err, setErr] = useState<string | null>(null);
	const [confirmRelId, setConfirmRelId] = useState<number | null>(null);

	const { data: rels } = useSuspenseQuery({
		queryKey: keys.people.relationships(personId),
		queryFn: () => listRelationships(personId),
	});
	const { data: types } = useSuspenseQuery({
		queryKey: keys.relationshipTypes.list(),
		queryFn: listRelationshipTypes,
	});
	const { data: personResults } = useQuery({
		queryKey: keys.people.list({ q: personSearch || undefined }),
		queryFn: () => listPeople({ q: personSearch || undefined, page_size: 10 }),
		enabled: personSearch.length > 0,
	});
	const { data: myProfile } = useSuspenseQuery({
		queryKey: keys.me.profile(),
		queryFn: getMe,
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
			qc.invalidateQueries({ queryKey: keys.relationships.graph(personId) });
			qc.invalidateQueries({ queryKey: keys.relationships.graph() });
			setAddOpen(false);
			setTypeId("");
			setOtherPersonId(null);
			setOtherPersonName("");
			setPersonSearch("");
			setNotes("");
		},
		onError: (e) =>
			setErr(apiErrorMessage(e, t("sections.relationships.failed"))),
	});

	const detach = useMutation({
		mutationFn: (relId: number) => detachRelationship(personId, relId),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.people.relationships(personId) });
			qc.invalidateQueries({ queryKey: keys.relationships.graph(personId) });
			qc.invalidateQueries({ queryKey: keys.relationships.graph() });
			setConfirmRelId(null);
		},
	});

	const confirmRel = rels.find((r) => r.id === confirmRelId);

	return (
		<div>
			<div className="flex items-center justify-between mb-2">
				<SectionHeading>{t("sections.relationships.title")}</SectionHeading>
				<div className="flex gap-1">
					<Button variant="neutral" size="sm" onClick={() => setBulkOpen(true)}>
						<ListPlus className="size-3" />{" "}
						{t("sections.relationships.addMultiple")}
					</Button>
					<Button variant="neutral" size="sm" onClick={() => setAddOpen(true)}>
						<Link2 className="size-3" /> {t("sections.relationships.add")}
					</Button>
				</div>
			</div>
			{!rels.length ? (
				<p className="text-sm text-sub">{t("sections.relationships.empty")}</p>
			) : (
				<div className="space-y-2">
					{rels.map((r) => (
						<div
							key={r.id}
							className="flex items-center gap-3 border-bw border-line rounded-md p-2 text-sm"
						>
							<RelationshipLabel
								display={display}
								typeName={r.type_name}
								inverseTypeName={r.inverse_type_name}
								selfName={personName}
								otherName={formatPersonName(
									r.other_person_name,
									r.other_person_nickname,
								)}
								other={
									<Link
										to="/people/$personId"
										params={{ personId: String(r.other_person_id) }}
										className="font-medium hover:underline"
									>
										{formatPersonName(
											r.other_person_name,
											r.other_person_nickname,
										)}
									</Link>
								}
							/>
							{r.notes && (
								<span className="text-sub text-xs truncate max-w-[140px]">
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
						<DialogTitle>{t("sections.relationships.addTitle")}</DialogTitle>
					</DialogHeader>
					{err && (
						<Alert variant="destructive">
							<AlertDescription>{err}</AlertDescription>
						</Alert>
					)}
					<div className="space-y-3">
						<div>
							<Label>{t("sections.relationships.type")}</Label>
							<select
								className="w-full h-10 border-field-bw border-field-line rounded-md bg-field px-2 text-sm"
								value={typeId}
								onChange={(e) => setTypeId(Number(e.target.value))}
							>
								<option value="">{t("sections.relationships.select")}</option>
								{types.map((rt) => (
									<option key={rt.id} value={rt.id}>
										{rt.name}
									</option>
								))}
							</select>
						</div>
						<div className="space-y-1">
							<div className="flex items-center gap-2">
								<Label>{t("sections.relationships.person")}</Label>
								{myProfile.id !== personId && !otherPersonId && (
									<button
										type="button"
										onClick={() => {
											setOtherPersonId(myProfile.id);
											setOtherPersonName(
												formatPersonName(myProfile.name, myProfile.nickname),
											);
											setPersonSearch("");
										}}
										className="text-xs text-accent-text hover:opacity-80 font-medium"
									>
										{t("sections.relationships.addMe")}
									</button>
								)}
							</div>
							{otherPersonId ? (
								<div className="flex items-center gap-2 border-bw border-line rounded-md px-3 py-2 text-sm">
									<span className="flex-1 font-medium">{otherPersonName}</span>
									<button
										type="button"
										onClick={() => {
											setOtherPersonId(null);
											setOtherPersonName("");
											setPersonSearch("");
										}}
										className="text-sub hover:text-destructive"
									>
										<X className="size-3" />
									</button>
								</div>
							) : (
								<div className="space-y-1">
									<div className="relative">
										<Search className="absolute left-2.5 top-2.5 size-3.5 text-sub" />
										<input
											type="text"
											value={personSearch}
											onChange={(e) => setPersonSearch(e.target.value)}
											placeholder={t(
												"sections.relationships.searchPlaceholder",
											)}
											className="h-9 w-full border-field-bw border-field-line rounded-md bg-field pl-8 pr-3 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
										/>
									</div>
									{personResults?.items && personResults.items.length > 0 && (
										<div className="border-bw border-line rounded-md bg-panel divide-y divide-line-soft max-h-40 overflow-y-auto">
											{personResults.items
												.filter((p) => p.id !== personId)
												.map((p) => (
													<button
														key={p.id}
														type="button"
														onClick={() => {
															setOtherPersonId(p.id);
															setOtherPersonName(
																formatPersonName(p.name, p.nickname),
															);
															setPersonSearch("");
														}}
														className="w-full text-left px-3 py-2 text-sm hover:bg-chip"
													>
														{formatPersonName(p.name, p.nickname)}
													</button>
												))}
										</div>
									)}
								</div>
							)}
						</div>
						<div>
							<Label>{t("sections.relationships.notes")}</Label>
							<Input
								value={notes}
								onChange={(e) => setNotes(e.target.value)}
								placeholder={t("sections.relationships.notesPlaceholder")}
							/>
						</div>
					</div>
					<DialogFooter>
						<Button variant="neutral" onClick={() => setAddOpen(false)}>
							{t("sections.relationships.cancel")}
						</Button>
						<Button
							disabled={attach.isPending || !typeId || !otherPersonId}
							onClick={() => attach.mutate()}
						>
							{t("sections.relationships.save")}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>

			<BulkRelationshipModal
				fromPersonId={personId}
				open={bulkOpen}
				onClose={() => setBulkOpen(false)}
				onSuccess={() => {
					qc.invalidateQueries({
						queryKey: keys.people.relationships(personId),
					});
					qc.invalidateQueries({
						queryKey: keys.relationships.graph(personId),
					});
					qc.invalidateQueries({ queryKey: keys.relationships.graph() });
				}}
			/>

			<Dialog
				open={confirmRelId !== null}
				onOpenChange={(v) => !v && setConfirmRelId(null)}
			>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>{t("sections.relationships.removeTitle")}</DialogTitle>
					</DialogHeader>
					{confirmRel && (
						<p className="text-[13px] text-sub">
							<SafeTrans
								t={t}
								i18nKey="sections.relationships.removeBody"
								values={{
									type: confirmRel.type_name,
									person: formatPersonName(
										confirmRel.other_person_name,
										confirmRel.other_person_nickname,
									),
								}}
								components={{
									type: <span className="font-medium" />,
									person: <span className="font-medium" />,
								}}
							/>
						</p>
					)}
					<DialogFooter>
						<Button variant="neutral" onClick={() => setConfirmRelId(null)}>
							{t("sections.relationships.cancel")}
						</Button>
						<Button
							variant="destructive"
							disabled={detach.isPending}
							onClick={() =>
								confirmRelId !== null && detach.mutate(confirmRelId)
							}
						>
							{detach.isPending
								? t("sections.relationships.removing")
								: t("sections.relationships.remove")}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</div>
	);
}
