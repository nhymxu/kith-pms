import { createFileRoute } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { PersonForm } from "#/features/people/person-form";

export const Route = createFileRoute("/_authed/people/new")({
	component: NewPersonPage,
});

function NewPersonPage() {
	const { t } = useTranslation("people");
	return (
		<div className="space-y-4">
			<h1 className="text-[18px] font-semibold tracking-tight text-ink font-display">
				{t("list.newPerson")}
			</h1>
			<PersonForm mode="create" />
		</div>
	);
}
