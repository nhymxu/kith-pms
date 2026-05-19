import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { Link } from "@tanstack/react-router"
import { useState } from "react"
import { Trash2, Plus, Link2, X } from "lucide-react"
import { Badge } from "#/components/ui/badge"
import { Button } from "#/components/ui/button"
import {
	Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter,
} from "#/components/ui/dialog"
import { Alert, AlertDescription } from "#/components/ui/alert"
import { Input } from "#/components/ui/input"
import { Label } from "#/components/ui/label"
import { keys } from "#/query-keys"
import { getPerson, listRelationships, attachRelationship, detachRelationship, attachLabel, detachLabel, listWorkHistory } from "#/endpoints/people"
import { listJournal } from "#/endpoints/journal"
import { listLabels } from "#/endpoints/labels"
import { listRelationshipTypes } from "#/endpoints/relationship-types"
import { listDatesByPerson } from "#/endpoints/dates"
import { listGifts } from "#/endpoints/gifts"
import type { Person } from "#/schemas/person"
import { AvatarUploader } from "./avatar-uploader"
import { QuickActions } from "./quick-actions"

function SectionHeading({ children }: { children: React.ReactNode }) {
	return <h2 className="text-[11px] font-semibold uppercase tracking-widest text-zinc-400 mb-2">{children}</h2>
}

function OverviewSection({ person }: { person: Person }) {
	return (
		<div className="space-y-3">
			<SectionHeading>Overview</SectionHeading>
			<AvatarUploader personId={person.id} hasAvatar={Boolean(person.avatar_path)} />
			<QuickActions personId={person.id} />
			<dl className="grid grid-cols-2 gap-x-4 gap-y-3 text-sm">
				{person.nickname && (
					<><dt className="font-medium text-zinc-500">Nickname</dt><dd>{person.nickname}</dd></>
				)}
				{person.relationship_type && (
					<><dt className="font-medium text-zinc-500">Relationship</dt><dd>{person.relationship_type}</dd></>
				)}
				{person.date_of_birth && (
					<><dt className="font-medium text-zinc-500">Date of birth</dt><dd>{person.date_of_birth}</dd></>
				)}
				{person.last_contact_at && (
					<><dt className="font-medium text-zinc-500">Last contact</dt><dd>{new Date(person.last_contact_at).toLocaleDateString()}</dd></>
				)}
			</dl>
			{person.other_notes && (
				<p className="text-sm font-base whitespace-pre-wrap border-l-2 border-border pl-3">{person.other_notes}</p>
			)}
		</div>
	)
}

function ContactsSection({ person }: { person: Person }) {
	return (
		<div>
			<SectionHeading>Contacts</SectionHeading>
			{!person.contacts?.length ? (
				<p className="text-sm text-zinc-400">No contacts.</p>
			) : (
				<div className="space-y-2">
					{person.contacts.map((c) => (
						<div key={c.id} className="flex gap-3 text-sm border border-zinc-200 rounded-md p-2">
							<Badge variant="neutral">{c.type}</Badge>
							<span className="font-base flex-1">{c.value}</span>
							{c.label && <span className="text-zinc-400">{c.label}</span>}
						</div>
					))}
				</div>
			)}
		</div>
	)
}

function LocationsSection({ person }: { person: Person }) {
	return (
		<div>
			<SectionHeading>Locations</SectionHeading>
			{!person.locations?.length ? (
				<p className="text-sm text-zinc-400">No locations.</p>
			) : (
				<div className="space-y-2">
					{person.locations.map((l) => (
						<div key={l.id} className="text-sm border border-zinc-200 rounded-md p-3 space-y-1">
							<Badge variant="neutral">{l.type}</Badge>
							<p className="font-base">{[l.address, l.city, l.country, l.postal_code].filter(Boolean).join(", ")}</p>
						</div>
					))}
				</div>
			)}
		</div>
	)
}

function LabelsSection({ person }: { person: Person }) {
	const qc = useQueryClient()
	const { data: allLabels } = useQuery({ queryKey: keys.labels.list(), queryFn: listLabels })
	const attached = person.labels ?? []
	const attachedIds = new Set(attached.map((l) => l.id))

	const attach = useMutation({
		mutationFn: (labelId: number) => attachLabel(person.id, labelId),
		onSuccess: () => qc.invalidateQueries({ queryKey: keys.people.detail(person.id) }),
	})
	const detach = useMutation({
		mutationFn: (labelId: number) => detachLabel(person.id, labelId),
		onSuccess: () => qc.invalidateQueries({ queryKey: keys.people.detail(person.id) }),
	})

	const available = allLabels?.filter((l) => !attachedIds.has(l.id)) ?? []

	return (
		<div>
			<SectionHeading>Labels</SectionHeading>
			<div className="space-y-2">
				<div className="flex flex-wrap gap-2">
					{attached.map((l) => (
						<div key={l.id} className="flex items-center gap-1">
							<Badge style={{ borderColor: l.color }}>{l.name}</Badge>
							<button type="button" className="text-foreground/40 hover:text-destructive" onClick={() => detach.mutate(l.id)}>
								<Trash2 className="size-3" />
							</button>
						</div>
					))}
					{attached.length === 0 && <p className="text-sm text-zinc-400">No labels attached.</p>}
				</div>
				{available.length > 0 && (
					<div className="flex flex-wrap gap-2">
						{available.map((l) => (
							<button
								key={l.id}
								type="button"
								onClick={() => attach.mutate(l.id)}
								className="flex items-center gap-1 text-xs border border-dashed border-zinc-300 rounded-md px-2 py-1 hover:border-main transition-colors"
							>
								<Plus className="size-3" />{l.name}
							</button>
						))}
					</div>
				)}
			</div>
		</div>
	)
}

function RelationshipsSection({ personId }: { personId: number }) {
	const qc = useQueryClient()
	const [addOpen, setAddOpen] = useState(false)
	const [typeId, setTypeId] = useState<number | "">("")
	const [otherId, setOtherId] = useState("")
	const [notes, setNotes] = useState("")
	const [err, setErr] = useState<string | null>(null)

	const { data: rels } = useQuery({
		queryKey: keys.people.relationships(personId),
		queryFn: () => listRelationships(personId),
	})
	const { data: types } = useQuery({
		queryKey: keys.relationshipTypes.list(),
		queryFn: listRelationshipTypes,
	})

	const attach = useMutation({
		mutationFn: () => attachRelationship(personId, {
			relationship_type_id: Number(typeId),
			to_person_id: Number(otherId),
			notes,
		}),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.people.relationships(personId) })
			setAddOpen(false); setTypeId(""); setOtherId(""); setNotes("")
		},
		onError: (e) => setErr(e instanceof Error ? e.message : "Failed"),
	})

	const detach = useMutation({
		mutationFn: (relId: number) => detachRelationship(personId, relId),
		onSuccess: () => qc.invalidateQueries({ queryKey: keys.people.relationships(personId) }),
	})

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
						<div key={r.id} className="flex items-center gap-3 border border-zinc-200 rounded-md p-2 text-sm">
							<Badge variant="neutral">{r.type_name}</Badge>
							<Link to="/people/$personId" params={{ personId: String(r.other_person_id) }} className="font-medium hover:underline flex-1">
								{r.other_person_name}
							</Link>
							{r.notes && <span className="text-zinc-400 text-xs">{r.notes}</span>}
							<button type="button" onClick={() => detach.mutate(r.id)} className="text-foreground/40 hover:text-destructive">
								<Trash2 className="size-3" />
							</button>
						</div>
					))}
				</div>
			)}

			<Dialog open={addOpen} onOpenChange={(v) => !v && setAddOpen(false)}>
				<DialogContent>
					<DialogHeader><DialogTitle>Add relationship</DialogTitle></DialogHeader>
					{err && <Alert variant="destructive"><AlertDescription>{err}</AlertDescription></Alert>}
					<div className="space-y-3">
						<div>
							<Label>Type</Label>
							<select
								className="w-full h-10 border border-zinc-200 rounded-md bg-white px-2 text-sm"
								value={typeId}
								onChange={(e) => setTypeId(Number(e.target.value))}
							>
								<option value="">Select…</option>
								{types?.map((t) => <option key={t.id} value={t.id}>{t.name}</option>)}
							</select>
						</div>
						<div><Label>Other person ID</Label><Input value={otherId} onChange={(e) => setOtherId(e.target.value)} placeholder="Person ID" /></div>
						<div><Label>Notes</Label><Input value={notes} onChange={(e) => setNotes(e.target.value)} placeholder="Optional notes" /></div>
					</div>
					<DialogFooter>
						<Button variant="neutral" onClick={() => setAddOpen(false)}>Cancel</Button>
						<Button disabled={attach.isPending || !typeId || !otherId} onClick={() => attach.mutate()}>Save</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</div>
	)
}

function WorkHistorySection({ personId }: { personId: number }) {
	const { data } = useQuery({
		queryKey: keys.people.workHistory(personId),
		queryFn: () => listWorkHistory(personId),
	})

	return (
		<div>
			<SectionHeading>Work History</SectionHeading>
			{!data?.length ? (
				<p className="text-sm text-zinc-400">No work history.</p>
			) : (
				<div className="space-y-3">
					{data.map((e) => (
						<div key={e.id} className="text-sm border border-zinc-200 rounded-md p-3 space-y-1">
							<p className="font-medium">{e.company}{e.title ? ` · ${e.title}` : ""}</p>
							<p className="font-mono text-[12px] text-zinc-500">
								{e.start_date} → {e.end_date || "Present"}
							</p>
							{e.location && <p className="text-zinc-500">{e.location}</p>}
							{e.description && <p className="text-zinc-600 whitespace-pre-wrap">{e.description}</p>}
						</div>
					))}
				</div>
			)}
		</div>
	)
}

function ImportantDatesSection({ personId }: { personId: number }) {
	const { data } = useQuery({
		queryKey: keys.dates.list(personId),
		queryFn: () => listDatesByPerson(personId),
	})

	return (
		<div>
			<SectionHeading>Important Dates</SectionHeading>
			{!data?.length ? (
				<p className="text-sm text-zinc-400">No important dates.</p>
			) : (
				<div className="space-y-2">
					{data.map((d) => (
						<div key={d.id} className="flex items-center gap-3 text-sm border border-zinc-200 rounded-md p-2">
							<Badge variant="neutral">{d.kind}</Badge>
							{d.label && <span className="text-zinc-500">{d.label}</span>}
							<span className="font-mono text-[12px] text-zinc-500 flex-1">{d.date_value}</span>
							{d.recurring && <span className="text-zinc-400 text-xs">↻</span>}
						</div>
					))}
				</div>
			)}
		</div>
	)
}

function GiftsSection({ personId }: { personId: number }) {
	const { data } = useQuery({
		queryKey: keys.gifts.list({ person_id: personId, page_size: 10 }),
		queryFn: () => listGifts({ person_id: personId, page_size: 10 }),
	})

	return (
		<div>
			<SectionHeading>Gifts</SectionHeading>
			{!data?.items?.length ? (
				<p className="text-sm text-zinc-400">No gifts.</p>
			) : (
				<div className="space-y-2">
					{data.items.map((g) => (
						<Link key={g.id} to="/gifts/$giftId" params={{ giftId: String(g.id) }} className="flex items-center gap-3 text-sm border border-zinc-200 rounded-md p-2 hover:bg-zinc-50">
							<span className="font-medium flex-1">{g.title}</span>
							<Badge variant="neutral">{g.direction}</Badge>
							{g.date && <span className="font-mono text-[12px] text-zinc-500">{g.date}</span>}
						</Link>
					))}
				</div>
			)}
		</div>
	)
}

function JournalSection({ personId }: { personId: number }) {
	const { data } = useQuery({
		queryKey: ["journal", "byPerson", personId],
		queryFn: () => listJournal({ page_size: 20 }),
	})
	const entries = data?.items.filter((e) => e.people.some((p) => p.person_id === personId)) ?? []

	return (
		<div>
			<div className="flex items-center justify-between mb-2">
				<SectionHeading>Journal</SectionHeading>
				<Button variant="neutral" size="sm" asChild>
					<Link to="/journal/new" search={{ person_id: personId }}>
						<Plus className="size-3" /> New entry
					</Link>
				</Button>
			</div>
			{entries.length === 0 ? (
				<p className="text-sm text-zinc-400">No journal entries for this person.</p>
			) : (
				<div className="space-y-2">
					{entries.map((e) => (
						<Link key={e.id} to="/journal/$entryId" params={{ entryId: String(e.id) }} className="block p-2 border border-zinc-200 rounded-md hover:underline text-sm">
							<span className="font-medium">{e.title}</span>
							<span className="text-zinc-400 ml-2">{e.occurred_at_date}</span>
						</Link>
					))}
				</div>
			)}
		</div>
	)
}

interface PersonDetailSectionsProps {
	personId: number
	onClose?: () => void
}

export function PersonDetailSections({ personId, onClose }: PersonDetailSectionsProps) {
	const { data: person, isLoading, error } = useQuery({
		queryKey: keys.people.detail(personId),
		queryFn: () => getPerson(personId),
	})

	if (isLoading) return <div className="py-12 text-center text-zinc-400 font-base">Loading…</div>
	if (error || !person) return <div className="py-12 text-center text-zinc-400 font-base">Person not found.</div>

	return (
		<div className="space-y-6">
			<div className="flex items-center justify-between">
				<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">{person.name}</h1>
				<div className="flex items-center gap-2">
					<Button variant="neutral" size="sm" asChild>
						<Link to="/people/$personId/edit" params={{ personId: String(personId) }}>Edit</Link>
					</Button>
					{onClose && (
						<Button variant="neutral" size="sm" onClick={onClose}>
							<X className="size-4" />
						</Button>
					)}
				</div>
			</div>

			<OverviewSection person={person} />
			<div className="border-t border-zinc-100" />
			<ContactsSection person={person} />
			<div className="border-t border-zinc-100" />
			<LocationsSection person={person} />
			<div className="border-t border-zinc-100" />
			<LabelsSection person={person} />
			<div className="border-t border-zinc-100" />
			<RelationshipsSection personId={personId} />
			<div className="border-t border-zinc-100" />
			<JournalSection personId={personId} />
			<div className="border-t border-zinc-100" />
			<WorkHistorySection personId={personId} />
			<div className="border-t border-zinc-100" />
			<ImportantDatesSection personId={personId} />
			<div className="border-t border-zinc-100" />
			<GiftsSection personId={personId} />
		</div>
	)
}
