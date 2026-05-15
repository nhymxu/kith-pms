import { createFileRoute, Link, useNavigate } from "@tanstack/react-router"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { useState } from "react"
import { getJournalEntry, deleteJournalEntry } from "#/endpoints/journal"
import { keys } from "#/query-keys"
import { Button } from "#/components/ui/button"
import { Badge } from "#/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "#/components/ui/card"
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
	DialogDescription,
	DialogFooter,
} from "#/components/ui/dialog"

export const Route = createFileRoute("/_authed/journal/$entryId")({
	component: JournalEntryPage,
})

function JournalEntryPage() {
	const { entryId } = Route.useParams()
	const id = Number(entryId)
	const navigate = useNavigate()
	const qc = useQueryClient()
	const [confirmOpen, setConfirmOpen] = useState(false)

	const { data, isPending, isError } = useQuery({
		queryKey: keys.journal.detail(id),
		queryFn: () => getJournalEntry(id),
	})

	const deleteMutation = useMutation({
		mutationFn: () => deleteJournalEntry(id),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.journal.all })
			navigate({ to: "/journal" })
		},
	})

	if (isPending) return <p className="text-sm font-base text-foreground/60">Loading…</p>
	if (isError || !data) return <p className="text-sm font-base text-destructive">Entry not found.</p>

	return (
		<div className="max-w-2xl space-y-4">
			<div className="flex items-center justify-between">
				<h1 className="text-2xl font-heading">{data.title}</h1>
				<div className="flex gap-2">
					<Button variant="neutral" asChild>
						<Link to="/journal/$entryId/edit" params={{ entryId }}>Edit</Link>
					</Button>
					<Button variant="destructive" onClick={() => setConfirmOpen(true)}>Delete</Button>
				</div>
			</div>

			<Card>
				<CardHeader>
					<CardTitle className="text-base font-heading text-foreground/60">
						{new Date(data.occurred_at_date).toLocaleDateString()}
						{data.occurred_at_time && ` at ${data.occurred_at_time}`}
					</CardTitle>
				</CardHeader>
				<CardContent className="space-y-3">
					{data.people.length > 0 && (
						<div className="flex flex-wrap gap-1">
							{data.people.map((p) => (
								<Badge key={p.person_id} variant="neutral">
									<Link to="/people/$personId" params={{ personId: String(p.person_id) }} className="hover:underline">
										{p.name}
									</Link>
								</Badge>
							))}
						</div>
					)}
					{data.content && (
						<p className="text-sm font-base whitespace-pre-wrap">{data.content}</p>
					)}
				</CardContent>
			</Card>

			<Dialog open={confirmOpen} onOpenChange={setConfirmOpen}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Delete journal entry?</DialogTitle>
						<DialogDescription>
							This will permanently delete "{data.title}". This action cannot be undone.
						</DialogDescription>
					</DialogHeader>
					<DialogFooter>
						<Button variant="neutral" onClick={() => setConfirmOpen(false)}>Cancel</Button>
						<Button
							variant="destructive"
							onClick={() => deleteMutation.mutate()}
							disabled={deleteMutation.isPending}
						>
							{deleteMutation.isPending ? "Deleting…" : "Delete"}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</div>
	)
}
