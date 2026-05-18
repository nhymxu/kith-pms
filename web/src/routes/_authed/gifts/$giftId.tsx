import { createFileRoute, Link, useNavigate } from "@tanstack/react-router"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { useState } from "react"
import { getGift, deleteGift } from "#/endpoints/gifts"
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

export const Route = createFileRoute("/_authed/gifts/$giftId")({
	component: GiftDetailPage,
})

function GiftDetailPage() {
	const { giftId } = Route.useParams()
	const id = Number(giftId)
	const navigate = useNavigate()
	const qc = useQueryClient()
	const [confirmOpen, setConfirmOpen] = useState(false)

	const { data, isPending, isError } = useQuery({
		queryKey: keys.gifts.detail(id),
		queryFn: () => getGift(id),
	})

	const deleteMutation = useMutation({
		mutationFn: () => deleteGift(id),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.gifts.all })
			navigate({ to: "/gifts" })
		},
	})

	if (isPending) return <p className="text-sm font-base text-foreground/60">Loading…</p>
	if (isError || !data) return <p className="text-sm font-base text-destructive">Gift not found.</p>

	return (
		<div className="max-w-lg space-y-4">
			<div className="flex items-center justify-between">
				<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">{data.title}</h1>
				<div className="flex gap-2">
					<Button variant="neutral" asChild>
						<Link to="/gifts/$giftId/edit" params={{ giftId }}>Edit</Link>
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
						<span className="text-foreground/60 w-28 shrink-0">Direction</span>
						<Badge variant="neutral">{data.direction}</Badge>
					</div>
					{data.debt_type && (
						<div className="flex gap-2">
							<span className="text-foreground/60 w-28 shrink-0">Debt</span>
							<span>{data.debt_type === "i_owe" ? "I owe" : "They owe"}</span>
						</div>
					)}
					{data.date && (
						<div className="flex gap-2">
							<span className="text-foreground/60 w-28 shrink-0">Date</span>
							<span>{new Date(data.date).toLocaleDateString()}</span>
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

			{/* Delete confirm dialog */}
			<Dialog open={confirmOpen} onOpenChange={setConfirmOpen}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Delete gift?</DialogTitle>
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
