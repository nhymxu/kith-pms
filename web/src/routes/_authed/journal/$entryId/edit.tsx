import {
	useMutation,
	useQueryClient,
	useSuspenseQuery,
} from "@tanstack/react-query";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { getJournalEntry, updateJournalEntry } from "#/endpoints/journal";
import { JournalForm } from "#/features/journal/journal-form";
import { keys } from "#/query-keys";
import type { JournalRequest } from "#/schemas/journal";

export const Route = createFileRoute("/_authed/journal/$entryId/edit")({
	component: EditJournalPage,
	pendingComponent: () => (
		<p className="text-sm font-base text-foreground/60">Loading…</p>
	),
	errorComponent: () => (
		<p className="text-sm font-base text-destructive">Entry not found.</p>
	),
});

function EditJournalPage() {
	const { entryId } = Route.useParams();
	const id = Number(entryId);
	const navigate = useNavigate();
	const qc = useQueryClient();

	const { data } = useSuspenseQuery({
		queryKey: keys.journal.detail(id),
		queryFn: () => getJournalEntry(id),
	});

	const mutation = useMutation({
		mutationFn: (body: JournalRequest) => updateJournalEntry(id, body),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.journal.detail(id) });
			qc.invalidateQueries({ queryKey: keys.journal.all });
			navigate({ to: "/journal/$entryId", params: { entryId } });
		},
	});

	return (
		<div className="space-y-4">
			<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">
				Edit Entry
			</h1>
			{/* key forces form remount once data is loaded so defaultValues are populated */}
			<JournalForm
				key={data.id}
				initial={data}
				onSubmit={(v) => mutation.mutateAsync(v)}
				submitLabel="Save Changes"
			/>
		</div>
	);
}
