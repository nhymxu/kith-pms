import {
	useMutation,
	useQueryClient,
	useSuspenseQuery,
} from "@tanstack/react-query";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { getJournalEntry, updateJournalEntry } from "#/endpoints/journal";
import { JournalForm } from "#/features/journal/journal-form";
import { keys } from "#/query-keys";
import type { JournalRequest } from "#/schemas/journal";

export const Route = createFileRoute("/_authed/journal/$entryId/edit")({
	component: EditJournalPage,
	pendingComponent: EditPending,
	errorComponent: EditError,
});

function EditPending() {
	const { t } = useTranslation("common");
	return <p className="text-sm font-base text-foreground/60">{t("loading")}</p>;
}

function EditError() {
	const { t } = useTranslation("journal");
	return (
		<p className="text-sm font-base text-destructive">{t("page.notFound")}</p>
	);
}

function EditJournalPage() {
	const { t } = useTranslation("journal");
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
			<h1 className="text-[18px] font-semibold tracking-tight text-ink font-display">
				{t("page.editTitle")}
			</h1>
			{/* key forces form remount once data is loaded so defaultValues are populated */}
			<JournalForm
				key={data.id}
				initial={data}
				onSubmit={(v) => mutation.mutateAsync(v)}
				submitLabel={t("form.saveChanges")}
			/>
		</div>
	);
}
