import {
	useMutation,
	useQueryClient,
	useSuspenseQuery,
} from "@tanstack/react-query";
import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { Badge } from "#/components/ui/badge";
import { Button } from "#/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "#/components/ui/card";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import { deleteGift, getGift } from "#/endpoints/gifts";
import { formatDate } from "#/lib/format-datetime";
import { keys } from "#/query-keys";

export const Route = createFileRoute("/_authed/gifts/$giftId/")({
	component: GiftDetailPage,
	pendingComponent: () => (
		<p className="text-sm font-base text-foreground/60">Loading…</p>
	),
	errorComponent: () => (
		<p className="text-sm font-base text-destructive">Gift not found.</p>
	),
});

function GiftDetailPage() {
	const { giftId } = Route.useParams();
	const id = Number(giftId);
	const navigate = useNavigate();
	const qc = useQueryClient();
	const [confirmOpen, setConfirmOpen] = useState(false);

	const { data } = useSuspenseQuery({
		queryKey: keys.gifts.detail(id),
		queryFn: () => getGift(id),
	});

	const deleteMutation = useMutation({
		mutationFn: () => deleteGift(id),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.gifts.all });
			navigate({ to: "/gifts" });
		},
	});

	const debtLabel =
		data.debt_type === "i_owe"
			? "I owe"
			: data.debt_type === "they_owe"
				? "They owe"
				: "—";

	const amountLabel =
		data.amount_cents != null
			? `${data.currency || "USD"} ${(data.amount_cents / 100).toFixed(2)}`
			: "—";

	return (
		<div className="max-w-lg space-y-4">
			<div className="flex items-center justify-between">
				<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">
					{data.title}
				</h1>
				<div className="flex gap-2">
					<Button variant="neutral" asChild>
						<Link to="/gifts/$giftId/edit" params={{ giftId }}>
							Edit
						</Link>
					</Button>
					<Button variant="destructive" onClick={() => setConfirmOpen(true)}>
						Delete
					</Button>
				</div>
			</div>

			<Card>
				<CardHeader>
					<CardTitle className="text-base">Details</CardTitle>
				</CardHeader>
				<CardContent className="space-y-2 text-sm font-base">
					<div className="flex gap-2">
						<span className="text-foreground/60 w-28 shrink-0">Person</span>
						<Link
							to="/people/$personId"
							params={{ personId: String(data.person_id) }}
							className="text-indigo-600 hover:underline"
						>
							{data.person_name}
						</Link>
					</div>
					<div className="flex gap-2">
						<span className="text-foreground/60 w-28 shrink-0">Direction</span>
						<Badge variant="neutral">{data.direction}</Badge>
					</div>
					<div className="flex gap-2">
						<span className="text-foreground/60 w-28 shrink-0">Debt</span>
						<span>{debtLabel}</span>
					</div>
					<div className="flex gap-2">
						<span className="text-foreground/60 w-28 shrink-0">Date</span>
						<span>{formatDate(data.date)}</span>
					</div>
					<div className="flex gap-2">
						<span className="text-foreground/60 w-28 shrink-0">Amount</span>
						<span>{amountLabel}</span>
					</div>
					<div className="flex gap-2">
						<span className="text-foreground/60 w-28 shrink-0">Notes</span>
						<span className="whitespace-pre-wrap">{data.notes || "—"}</span>
					</div>
					<div className="flex gap-2">
						<span className="text-foreground/60 w-28 shrink-0">Image</span>
						{data.image_path ? (
							<img
								src={`/v1/gifts/${data.id}/image`}
								alt={data.title}
								className="mt-1 rounded border border-zinc-200 max-h-64 object-contain"
							/>
						) : (
							<span className="text-zinc-400">No image</span>
						)}
					</div>
				</CardContent>
			</Card>

			<Dialog open={confirmOpen} onOpenChange={setConfirmOpen}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Delete gift?</DialogTitle>
						<DialogDescription>
							This will permanently delete "{data.title}". This action cannot be
							undone.
						</DialogDescription>
					</DialogHeader>
					<DialogFooter>
						<Button variant="neutral" onClick={() => setConfirmOpen(false)}>
							Cancel
						</Button>
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
	);
}
