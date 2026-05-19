import { createFileRoute, Link, useNavigate } from "@tanstack/react-router"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { useState } from "react"
import { getReminder, deleteReminder } from "#/endpoints/reminders"
import { formatDate } from "#/lib/format-datetime"
import { keys } from "#/query-keys"
import { CompleteButton } from "#/features/reminders/complete-button"
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

export const Route = createFileRoute("/_authed/reminders/$reminderId")({
	component: ReminderDetailPage,
})

function ReminderDetailPage() {
	const { reminderId } = Route.useParams()
	const id = Number(reminderId)
	const navigate = useNavigate()
	const qc = useQueryClient()
	const [confirmOpen, setConfirmOpen] = useState(false)

	const { data, isPending, isError } = useQuery({
		queryKey: keys.reminders.detail(id),
		queryFn: () => getReminder(id),
	})

	const deleteMutation = useMutation({
		mutationFn: () => deleteReminder(id),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.reminders.all })
			navigate({ to: "/reminders" })
		},
	})

	if (isPending) return <p className="text-sm font-base text-foreground/60">Loading…</p>
	if (isError || !data) return <p className="text-sm font-base text-destructive">Reminder not found.</p>

	const isOverdue = !data.completed && new Date(data.due_date) < new Date()

	return (
		<div className="max-w-lg space-y-4">
			<div className="flex items-center justify-between">
				<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">{data.title}</h1>
				<div className="flex gap-2">
					{!data.completed && <CompleteButton reminderId={id} />}
					<Button variant="neutral" asChild>
						<Link to="/reminders/$reminderId/edit" params={{ reminderId }}>Edit</Link>
					</Button>
					<Button variant="destructive" onClick={() => setConfirmOpen(true)}>Delete</Button>
				</div>
			</div>

			<Card>
				<CardHeader>
					<CardTitle className="text-base">Details</CardTitle>
				</CardHeader>
				<CardContent className="space-y-2 text-sm font-base">
					<div className="flex gap-2">
						<span className="text-foreground/60 w-28 shrink-0">Status</span>
						{data.completed
							? <Badge variant="neutral">Done</Badge>
							: isOverdue
								? <Badge className="bg-red-300 text-black border-black">Overdue</Badge>
								: <Badge>Upcoming</Badge>}
					</div>
					<div className="flex gap-2">
						<span className="text-foreground/60 w-28 shrink-0">Due</span>
						<span>{formatDate(data.due_date)}</span>
					</div>
					{data.person_name && (
						<div className="flex gap-2">
							<span className="text-foreground/60 w-28 shrink-0">Person</span>
							<span>{data.person_name}</span>
						</div>
					)}
					{data.notes && (
						<div className="flex gap-2">
							<span className="text-foreground/60 w-28 shrink-0">Notes</span>
							<span className="whitespace-pre-wrap">{data.notes}</span>
						</div>
					)}
				</CardContent>
			</Card>

			<Dialog open={confirmOpen} onOpenChange={setConfirmOpen}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Delete reminder?</DialogTitle>
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
