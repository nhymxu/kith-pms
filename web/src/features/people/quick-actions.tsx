import { useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { BookOpen, Gift, Clock } from "lucide-react"
import { Button } from "#/components/ui/button"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "#/components/ui/dialog"
import { Input } from "#/components/ui/input"
import { Label } from "#/components/ui/label"
import { Textarea } from "#/components/ui/textarea"
import { Alert, AlertDescription } from "#/components/ui/alert"
import { SubmitButton } from "#/components/form/submit-button"
import { apiFetch } from "#/lib/api-client"
import { keys } from "#/query-keys"

interface QuickActionsProps {
	personId: number
}

export function QuickJournalDialog({ personId, open, onClose }: { personId: number; open: boolean; onClose: () => void }) {
	const [title, setTitle] = useState("")
	const [content, setContent] = useState("")
	const [error, setError] = useState<string | null>(null)
	const qc = useQueryClient()

	const mutation = useMutation({
		mutationFn: () =>
			apiFetch(`/v1/people/${personId}/journal/quick`, {
				method: "POST",
				body: JSON.stringify({ title, content }),
			}),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.journal.all })
			qc.invalidateQueries({ queryKey: keys.people.detail(personId) })
			setTitle(""); setContent(""); onClose()
		},
		onError: (err) => setError(err instanceof Error ? err.message : "Failed"),
	})

	return (
		<Dialog open={open} onOpenChange={(v) => !v && onClose()}>
			<DialogContent>
				<DialogHeader><DialogTitle>Quick journal entry</DialogTitle></DialogHeader>
				{error && <Alert variant="destructive"><AlertDescription>{error}</AlertDescription></Alert>}
				<div className="space-y-3">
					<div><Label>Title</Label><Input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="What happened?" /></div>
					<div><Label>Notes</Label><Textarea value={content} onChange={(e) => setContent(e.target.value)} rows={3} placeholder="Details…" /></div>
				</div>
				<DialogFooter>
					<Button variant="neutral" onClick={onClose}>Cancel</Button>
					<SubmitButton isPending={mutation.isPending} onClick={() => mutation.mutate()} type="button">Save</SubmitButton>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	)
}

export function QuickGiftDialog({ personId, open, onClose }: { personId: number; open: boolean; onClose: () => void }) {
	const [title, setTitle] = useState("")
	const [error, setError] = useState<string | null>(null)
	const qc = useQueryClient()

	const mutation = useMutation({
		mutationFn: () =>
			apiFetch(`/v1/people/${personId}/gifts/quick`, {
				method: "POST",
				body: JSON.stringify({ title }),
			}),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.gifts.all })
			setTitle(""); onClose()
		},
		onError: (err) => setError(err instanceof Error ? err.message : "Failed"),
	})

	return (
		<Dialog open={open} onOpenChange={(v) => !v && onClose()}>
			<DialogContent>
				<DialogHeader><DialogTitle>Quick gift note</DialogTitle></DialogHeader>
				{error && <Alert variant="destructive"><AlertDescription>{error}</AlertDescription></Alert>}
				<div><Label>Gift title</Label><Input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="e.g. Birthday gift idea" /></div>
				<DialogFooter>
					<Button variant="neutral" onClick={onClose}>Cancel</Button>
					<SubmitButton isPending={mutation.isPending} onClick={() => mutation.mutate()} type="button">Save</SubmitButton>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	)
}

export function QuickActions({ personId }: QuickActionsProps) {
	const [journalOpen, setJournalOpen] = useState(false)
	const [giftOpen, setGiftOpen] = useState(false)
	const qc = useQueryClient()

	const lastContactMutation = useMutation({
		mutationFn: () =>
			apiFetch(`/v1/people/${personId}/last-contact`, { method: "POST" }),
		onSuccess: () => qc.invalidateQueries({ queryKey: keys.people.detail(personId) }),
	})

	return (
		<>
			<div className="flex flex-wrap gap-2">
				<Button variant="neutral" size="sm" onClick={() => setJournalOpen(true)}>
					<BookOpen className="size-3" /> Quick journal
				</Button>
				<Button variant="neutral" size="sm" onClick={() => setGiftOpen(true)}>
					<Gift className="size-3" /> Quick gift
				</Button>
				<Button
					variant="neutral"
					size="sm"
					disabled={lastContactMutation.isPending}
					onClick={() => lastContactMutation.mutate()}
				>
					<Clock className="size-3" /> Update last contact
				</Button>
			</div>

			<QuickJournalDialog personId={personId} open={journalOpen} onClose={() => setJournalOpen(false)} />
			<QuickGiftDialog personId={personId} open={giftOpen} onClose={() => setGiftOpen(false)} />
		</>
	)
}
