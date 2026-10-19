import { createFileRoute } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { DatesList } from "#/features/important_dates/dates-list";

export const Route = createFileRoute("/_authed/important-dates/")({
	component: DatesPage,
});

function DatesPage() {
	const { t } = useTranslation("dates");
	return (
		<div className="space-y-4">
			<h1 className="text-[18px] font-semibold tracking-tight text-ink font-display">
				{t("page.title")}
			</h1>
			<DatesList />
		</div>
	);
}
