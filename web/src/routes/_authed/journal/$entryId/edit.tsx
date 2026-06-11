import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { getJournalEntry, updateJournalEntry } from "#/endpoints/journal";
import { JournalForm } from "#/features/journal/journal-form";
import { keys } from "#/query-keys";
import type { JournalRequest } from "#/schemas/journal";

export const Route = createFileRoute("/_authed/journal/$entryId/edit")({
	component: EditJournalPage,
});

function EditJournalPage() {
	const { entryId } = Route.useParams();
	const id = Number(entryId);
	const navigate = useNavigate();
	const qc = useQueryClient();

	const { data, isPending, isError } = useQuery({
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

	if (isPending)
		return <p className="text-sm font-base text-foreground/60">Loading…</p>;
	if (isError || !data)
		return (
			<p className="text-sm font-base text-destructive">Entry not found.</p>
		);

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
