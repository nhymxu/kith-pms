import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { useState } from "react"
import { Pencil, Trash2, Plus, Check, X } from "lucide-react"
import { Button } from "#/components/ui/button"
import { Input } from "#/components/ui/input"
import { Textarea } from "#/components/ui/textarea"
import { Alert, AlertDescription } from "#/components/ui/alert"
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "#/components/ui/dialog"
import { keys } from "#/query-keys"
import { listWorkHistory, replaceWorkHistory } from "#/endpoints/people"
import type { WorkEntry } from "#/schemas/work-history"

function SectionHeading({ children }: { children: React.ReactNode }) {
	return <h2 className="text-[11px] font-semibold uppercase tracking-widest text-zinc-400 mb-2">{children}</h2>
}

interface EntryFormProps {
	entry: Partial<WorkEntry>
	onSave: (e: Partial<WorkEntry>) => void
	onCancel: () => void
}

function WorkEntryForm({ entry, onSave, onCancel }: EntryFormProps) {
	const [company, setCompany] = useState(entry.company ?? "")
	const [title, setTitle] = useState(entry.title ?? "")
	const [startDate, setStartDate] = useState(entry.start_date ?? "")
	const [endDate, setEndDate] = useState(entry.end_date ?? "")
	const [location, setLocation] = useState(entry.location ?? "")
	const [description, setDescription] = useState(entry.description ?? "")

	return (
		<div className="border border-zinc-300 rounded-md p-3 bg-zinc-50 space-y-2">
			<div className="flex gap-2">
				<Input className="h-8 flex-1" value={company} onChange={(e) => setCompany(e.target.value)} placeholder="Company *" />
				<Input className="h-8 flex-1" value={title} onChange={(e) => setTitle(e.target.value)} placeholder="Title" />
			</div>
			<div className="flex gap-2">
				<input type="month" className="h-8 flex-1 rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm" value={startDate} onChange={(e) => setStartDate(e.target.value)} placeholder="Start date" />
				<input type="month" className="h-8 flex-1 rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm" value={endDate} onChange={(e) => setEndDate(e.target.value)} placeholder="End date (blank = Present)" />
			</div>
			<Input className="h-8" value={location} onChange={(e) => setLocation(e.target.value)} placeholder="Location" />
			<Textarea rows={2} value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Description" />
			<div className="flex gap-2 justify-end">
				<Button variant="neutral" size="sm" onClick={onCancel}><X className="size-3" /> Cancel</Button>
				<Button size="sm" disabled={!company} onClick={() => onSave({ ...entry, company, title, start_date: startDate, end_date: endDate, location, description })}>
					<Check className="size-3" /> Save
				</Button>
			</div>
		</div>
	)
}

interface WorkHistorySectionProps {
	personId: number
}

export function WorkHistorySection({ personId }: WorkHistorySectionProps) {
	const qc = useQueryClient()
	const [editingId, setEditingId] = useState<number | "new" | null>(null)
	const [saveError, setSaveError] = useState<string | null>(null)
	const [confirmDeleteId, setConfirmDeleteId] = useState<number | null>(null)

	const { data = [] } = useQuery({
		queryKey: keys.people.workHistory(personId),
		queryFn: () => listWorkHistory(personId),
	})

	const saveMutation = useMutation({
		mutationFn: (entries: WorkEntry[]) =>
			replaceWorkHistory(personId, {
				entries: entries.map((e, i) => ({
					company: e.company,
					title: e.title ?? "",
					start_date: e.start_date,
					end_date: e.end_date ?? "",
					location: e.location ?? "",
					description: e.description ?? "",
					position: i,
				})),
			}),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.people.workHistory(personId) })
			setEditingId(null)
			setSaveError(null)
		},
		onError: (e) => setSaveError(e instanceof Error ? e.message : "Failed to save"),
	})

	function handleSave(updated: Partial<WorkEntry>) {
		let entries: WorkEntry[]
		if (editingId === "new") {
			entries = [...data, { id: 0, person_id: personId, created_at: "", position: data.length, company: updated.company!, title: updated.title ?? "", start_date: updated.start_date ?? "", end_date: updated.end_date ?? "", location: updated.location ?? "", description: updated.description ?? "" }]
		} else {
			entries = data.map((e) => e.id === editingId ? { ...e, ...updated } : e)
		}
		saveMutation.mutate(entries)
	}

	function handleDelete(id: number) {
		saveMutation.mutate(data.filter((e) => e.id !== id))
	}

	return (
		<div>
			<div className="flex items-center justify-between mb-2">
				<SectionHeading>Work History</SectionHeading>
				<Button variant="neutral" size="sm" onClick={() => setEditingId("new")}>
					<Plus className="size-3" /> Add
				</Button>
			</div>
			{saveError && <Alert variant="destructive" className="mb-2"><AlertDescription>{saveError}</AlertDescription></Alert>}
			<div className="space-y-3">
			</div>
			<div className="space-y-3">
				{data.map((e) =>
					editingId === e.id ? (
						<WorkEntryForm key={e.id} entry={e} onSave={handleSave} onCancel={() => setEditingId(null)} />
					) : (
						<div key={e.id} className="text-sm border border-zinc-200 rounded-md p-3 space-y-1">
							<div className="flex items-start gap-2">
								<p className="font-medium flex-1">{e.company}{e.title ? ` · ${e.title}` : ""}</p>
								<button type="button" onClick={() => setEditingId(e.id)} className="text-foreground/40 hover:text-main">
									<Pencil className="size-3" />
								</button>
								<button type="button" onClick={() => setConfirmDeleteId(e.id)} className="text-foreground/40 hover:text-destructive">
									<Trash2 className="size-3" />
								</button>
							</div>
							<p className="font-mono text-[12px] text-zinc-500">{e.start_date} → {e.end_date || "Present"}</p>
							{e.location && <p className="text-zinc-500">{e.location}</p>}
							{e.description && <p className="text-zinc-600 whitespace-pre-wrap">{e.description}</p>}
						</div>
					)
				)}
				{editingId === "new" && (
					<WorkEntryForm entry={{}} onSave={handleSave} onCancel={() => setEditingId(null)} />
				)}
				{data.length === 0 && editingId !== "new" && (
					<p className="text-sm text-zinc-400">No work history.</p>
				)}
			</div>
			<Dialog open={confirmDeleteId !== null} onOpenChange={(v) => !v && setConfirmDeleteId(null)}>
				<DialogContent>
					<DialogHeader><DialogTitle>Remove work entry?</DialogTitle></DialogHeader>
					{(() => {
						const e = data.find((e) => e.id === confirmDeleteId)
						return e ? (
							<p className="text-[13px] text-zinc-600">
								Remove <span className="font-medium">{e.company}</span>{e.title ? ` · ${e.title}` : ""}?
							</p>
						) : null
					})()}
					<DialogFooter>
						<Button variant="neutral" onClick={() => setConfirmDeleteId(null)}>Cancel</Button>
						<Button
							variant="destructive"
							disabled={saveMutation.isPending}
							onClick={() => { if (confirmDeleteId !== null) { handleDelete(confirmDeleteId); setConfirmDeleteId(null) } }}
						>
							{saveMutation.isPending ? "Removing…" : "Remove"}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</div>
	)
}
