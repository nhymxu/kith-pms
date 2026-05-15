import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { Link } from "@tanstack/react-router"
import { useState } from "react"
import { Trash2, Plus, Link2 } from "lucide-react"
import { Tabs, TabsList, TabsTrigger, TabsContent } from "#/components/ui/tabs"
import { Card, CardContent } from "#/components/ui/card"
import { Badge } from "#/components/ui/badge"
import { Button } from "#/components/ui/button"
import {
	Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter,
} from "#/components/ui/dialog"
import { Alert, AlertDescription } from "#/components/ui/alert"
import { Input } from "#/components/ui/input"
import { Label } from "#/components/ui/label"
import { keys } from "#/query-keys"
import { getPerson, listRelationships, attachRelationship, detachRelationship, attachLabel, detachLabel } from "#/endpoints/people"
import { listJournal } from "#/endpoints/journal"
import { listLabels } from "#/endpoints/labels"
import { listRelationshipTypes } from "#/endpoints/relationship-types"
import type { Person } from "#/schemas/person"
import { AvatarUploader } from "./avatar-uploader"
import { QuickActions } from "./quick-actions"

// ── Overview tab ────────────────────────────────────────────────────────────

function OverviewTab({ person }: { person: Person }) {
	return (
		<div className="space-y-4">
			<AvatarUploader personId={person.id} hasAvatar={Boolean(person.avatar_path)} />
			<QuickActions personId={person.id} />
			<dl className="grid grid-cols-2 gap-x-4 gap-y-3 text-sm">
				{person.nickname && (
					<><dt className="font-heading text-foreground/60">Nickname</dt><dd>{person.nickname}</dd></>
				)}
				{person.relationship_type && (
					<><dt className="font-heading text-foreground/60">Relationship</dt><dd>{person.relationship_type}</dd></>
				)}
				{person.date_of_birth && (
					<><dt className="font-heading text-foreground/60">Date of birth</dt><dd>{person.date_of_birth}</dd></>
				)}
				{person.last_contact_at && (
					<><dt className="font-heading text-foreground/60">Last contact</dt><dd>{new Date(person.last_contact_at).toLocaleDateString()}</dd></>
				)}
			</dl>
			{person.other_notes && (
				<p className="text-sm font-base whitespace-pre-wrap border-l-2 border-border pl-3">{person.other_notes}</p>
			)}
		</div>
	)
}

// ── Contacts tab ─────────────────────────────────────────────────────────────

function ContactsTab({ person }: { person: Person }) {
	if (!person.contacts?.length) return <p className="text-sm text-foreground/50">No contacts.</p>
	return (
		<div className="space-y-2">
			{person.contacts.map((c) => (
				<div key={c.id} className="flex gap-3 text-sm border-2 border-border rounded-base p-2">
					<Badge variant="neutral">{c.type}</Badge>
					<span className="font-base flex-1">{c.value}</span>
					{c.label && <span className="text-foreground/50">{c.label}</span>}
				</div>
			))}
		</div>
	)
}

// ── Locations tab ─────────────────────────────────────────────────────────────

function LocationsTab({ person }: { person: Person }) {
	if (!person.locations?.length) return <p className="text-sm text-foreground/50">No locations.</p>
	return (
		<div className="space-y-2">
			{person.locations.map((l) => (
				<div key={l.id} className="text-sm border-2 border-border rounded-base p-3 space-y-1">
					<Badge variant="neutral">{l.type}</Badge>
					<p className="font-base">{[l.address, l.city, l.country, l.postal_code].filter(Boolean).join(", ")}</p>
				</div>
			))}
		</div>
	)
}

// ── Labels tab ───────────────────────────────────────────────────────────────

function LabelsTab({ person }: { person: Person }) {
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

	return (
		<div className="space-y-3">
			<div className="flex flex-wrap gap-2">
				{attached.map((l) => (
					<div key={l.id} className="flex items-center gap-1">
						<Badge style={{ borderColor: l.color }}>{l.name}</Badge>
						<button
							type="button"
							className="text-foreground/40 hover:text-destructive"
							onClick={() => detach.mutate(l.id)}
						>
							<Trash2 className="size-3" />
						</button>
					</div>
				))}
				{attached.length === 0 && <p className="text-sm text-foreground/50">No labels attached.</p>}
			</div>
			{allLabels && allLabels.filter((l) => !attachedIds.has(l.id)).length > 0 && (
				<div>
					<p className="text-xs font-heading text-foreground/60 mb-1">Add label</p>
					<div className="flex flex-wrap gap-2">
						{allLabels.filter((l) => !attachedIds.has(l.id)).map((l) => (
							<button
								key={l.id}
								type="button"
								onClick={() => attach.mutate(l.id)}
								className="flex items-center gap-1 text-xs border-2 border-dashed border-border rounded-base px-2 py-1 hover:border-main transition-colors"
							>
								<Plus className="size-3" />{l.name}
							</button>
						))}
					</div>
				</div>
			)}
		</div>
	)
}

// ── Relationships tab ─────────────────────────────────────────────────────────

function RelationshipsTab({ personId }: { personId: number }) {
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
		mutationFn: () =>
			attachRelationship(personId, {
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
		<div className="space-y-3">
			<Button variant="neutral" size="sm" onClick={() => setAddOpen(true)}>
				<Link2 className="size-3" /> Add relationship
			</Button>
			{rels?.map((r) => (
				<div key={r.id} className="flex items-center gap-3 border-2 border-border rounded-base p-2 text-sm">
					<Badge variant="neutral">{r.type_name}</Badge>
					<Link to="/people/$personId" params={{ personId: String(r.other_person_id) }} className="font-heading hover:underline flex-1">
						{r.other_person_name}
					</Link>
					{r.notes && <span className="text-foreground/50 text-xs">{r.notes}</span>}
					<button type="button" onClick={() => detach.mutate(r.id)} className="text-foreground/40 hover:text-destructive">
						<Trash2 className="size-3" />
					</button>
				</div>
			))}
			{!rels?.length && <p className="text-sm text-foreground/50">No relationships yet.</p>}

			<Dialog open={addOpen} onOpenChange={(v) => !v && setAddOpen(false)}>
				<DialogContent>
					<DialogHeader><DialogTitle>Add relationship</DialogTitle></DialogHeader>
					{err && <Alert variant="destructive"><AlertDescription>{err}</AlertDescription></Alert>}
					<div className="space-y-3">
						<div>
							<Label>Type</Label>
							<select
								className="w-full h-10 border-2 border-border rounded-base bg-background px-2 text-sm"
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

// ── Journal tab ───────────────────────────────────────────────────────────────

function JournalTab({ personId }: { personId: number }) {
	const { data } = useQuery({
		queryKey: ["journal", "byPerson", personId],
		queryFn: () => listJournal({ page_size: 20 }),
	})
	const entries = data?.items.filter((e) => e.people.some((p) => p.person_id === personId)) ?? []

	return (
		<div className="space-y-2">
			<Button variant="neutral" size="sm" asChild>
				<Link to="/journal/new" search={{ person_id: personId }}>
					<Plus className="size-3" /> New entry
				</Link>
			</Button>
			{entries.length === 0 && <p className="text-sm text-foreground/50">No journal entries for this person.</p>}
			{entries.map((e) => (
				<Link key={e.id} to="/journal/$entryId" params={{ entryId: String(e.id) }} className="block p-2 border-2 border-border rounded-base hover:underline text-sm">
					<span className="font-heading">{e.title}</span>
					<span className="text-foreground/50 ml-2">{e.occurred_at_date}</span>
				</Link>
			))}
		</div>
	)
}

// ── Main export ───────────────────────────────────────────────────────────────

interface PersonDetailSectionsProps {
	personId: number
}

export function PersonDetailSections({ personId }: PersonDetailSectionsProps) {
	const { data: person, isLoading, error } = useQuery({
		queryKey: keys.people.detail(personId),
		queryFn: () => getPerson(personId),
	})

	if (isLoading) return <div className="py-12 text-center text-foreground/50 font-base">Loading…</div>
	if (error || !person) return <div className="py-12 text-center text-foreground/50 font-base">Person not found.</div>

	return (
		<div className="space-y-4">
			<div className="flex items-center justify-between">
				<h1 className="text-2xl font-heading">{person.name}</h1>
				<Button variant="neutral" size="sm" asChild>
					<Link to="/people/$personId/edit" params={{ personId: String(personId) }}>Edit</Link>
				</Button>
			</div>

			<Tabs defaultValue="overview">
				<TabsList>
					<TabsTrigger value="overview">Overview</TabsTrigger>
					<TabsTrigger value="contacts">Contacts</TabsTrigger>
					<TabsTrigger value="locations">Locations</TabsTrigger>
					<TabsTrigger value="labels">Labels</TabsTrigger>
					<TabsTrigger value="relationships">Relationships</TabsTrigger>
					<TabsTrigger value="journal">Journal</TabsTrigger>
				</TabsList>

				<TabsContent value="overview">
					<Card><CardContent className="pt-4"><OverviewTab person={person} /></CardContent></Card>
				</TabsContent>
				<TabsContent value="contacts">
					<Card><CardContent className="pt-4"><ContactsTab person={person} /></CardContent></Card>
				</TabsContent>
				<TabsContent value="locations">
					<Card><CardContent className="pt-4"><LocationsTab person={person} /></CardContent></Card>
				</TabsContent>
				<TabsContent value="labels">
					<Card><CardContent className="pt-4"><LabelsTab person={person} /></CardContent></Card>
				</TabsContent>
				<TabsContent value="relationships">
					<Card><CardContent className="pt-4"><RelationshipsTab personId={personId} /></CardContent></Card>
				</TabsContent>
				<TabsContent value="journal">
					<Card><CardContent className="pt-4"><JournalTab personId={personId} /></CardContent></Card>
				</TabsContent>
			</Tabs>
		</div>
	)
}
