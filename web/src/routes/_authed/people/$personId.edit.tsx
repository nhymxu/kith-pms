import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { getPerson } from "#/endpoints/people";
import { PersonForm } from "#/features/people/person-form";
import { keys } from "#/query-keys";

export const Route = createFileRoute("/_authed/people/$personId/edit")({
	component: EditPersonPage,
	pendingComponent: EditPending,
	errorComponent: EditError,
});

function EditPending() {
	const { t } = useTranslation("people");
	return (
		<p className="text-sm font-base text-foreground/60">{t("edit.loading")}</p>
	);
}

function EditError() {
	const { t } = useTranslation("people");
	return (
		<p className="text-sm font-base text-destructive">{t("edit.notFound")}</p>
	);
}

function EditPersonPage() {
	const { t } = useTranslation("people");
	const { personId } = Route.useParams();
	const id = Number(personId);

	const { data } = useSuspenseQuery({
		queryKey: keys.people.detail(id),
		queryFn: () => getPerson(id),
	});

	return (
		<div className="space-y-4">
			<h1 className="text-[18px] font-semibold tracking-tight text-ink font-display">
				{t("edit.title", { name: data.name })}
			</h1>
			<PersonForm mode="edit" initial={data} />
		</div>
	);
}
