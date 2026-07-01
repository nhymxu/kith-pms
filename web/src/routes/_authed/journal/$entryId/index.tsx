import {
	useMutation,
	useQueryClient,
	useSuspenseQuery,
} from "@tanstack/react-query";
import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
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
import { deleteJournalEntry, getJournalEntry } from "#/endpoints/journal";
import { LabelChip, PersonChip } from "#/features/journal/person-label-chip";
import { formatDate, formatTime } from "#/lib/format-datetime";
import { keys } from "#/query-keys";

export const Route = createFileRoute("/_authed/journal/$entryId/")({
	component: JournalEntryPage,
	pendingComponent: () => (
		<p className="text-sm font-base text-foreground/60">Loading…</p>
	),
	errorComponent: () => (
		<p className="text-sm font-base text-destructive">Entry not found.</p>
	),
});

function JournalEntryPage() {
	const { entryId } = Route.useParams();
	const id = Number(entryId);
	const navigate = useNavigate();
	const qc = useQueryClient();
	const [confirmOpen, setConfirmOpen] = useState(false);

	const { data } = useSuspenseQuery({
		queryKey: keys.journal.detail(id),
		queryFn: () => getJournalEntry(id),
	});

	const deleteMutation = useMutation({
		mutationFn: () => deleteJournalEntry(id),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.journal.all });
			navigate({ to: "/journal" });
		},
	});

	return (
		<div className="max-w-[760px] space-y-4">
			<div className="flex items-center justify-between">
				<div className="flex items-center gap-3">
					<Link
						to="/journal"
						className="text-[12px] text-zinc-400 hover:text-zinc-700"
					>
						← Journal
					</Link>
					<h1 className="text-[20px] font-semibold tracking-tight text-zinc-900">
						{data.title}
					</h1>
				</div>
				<div className="flex gap-2">
					<Button variant="neutral" asChild>
						<Link to="/journal/$entryId/edit" params={{ entryId }}>
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
					<CardTitle className="font-mono text-[12px] text-zinc-500">
						{formatDate(data.occurred_at_date)}
						{data.occurred_at_time
							? ` at ${formatTime(data.occurred_at_time)}`
							: ""}
					</CardTitle>
				</CardHeader>
				<CardContent className="space-y-3">
					{data.labels.length > 0 && (
						<div>
							<p className="text-[11px] font-semibold uppercase tracking-widest text-zinc-400 mb-1">
								Labels
							</p>
							<div className="flex flex-wrap gap-1.5">
								{data.labels.map((l) => (
									<LabelChip key={l.id} label={l} />
								))}
							</div>
						</div>
					)}
					<div>
						<p className="text-[11px] font-semibold uppercase tracking-widest text-zinc-400 mb-1">
							People
						</p>
						{data.people.length > 0 ? (
							<div className="flex flex-wrap gap-1.5">
								{data.people.map((p) => (
									<PersonChip key={p.person_id} p={p} />
								))}
							</div>
						) : (
							<p className="text-[13px] text-zinc-400">No people linked.</p>
						)}
					</div>
					<div>
						<p className="text-[11px] font-semibold uppercase tracking-widest text-zinc-400 mb-1">
							Notes
						</p>
						{data.content ? (
							<p className="text-[13px] text-zinc-700 whitespace-pre-wrap leading-relaxed">
								{data.content}
							</p>
						) : (
							<p className="text-[13px] text-zinc-400">No notes.</p>
						)}
					</div>
				</CardContent>
			</Card>

			<Dialog open={confirmOpen} onOpenChange={setConfirmOpen}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Delete journal entry?</DialogTitle>
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
