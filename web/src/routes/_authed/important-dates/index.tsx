import { createFileRoute } from "@tanstack/react-router";
import { DatesList } from "#/features/important_dates/dates-list";

export const Route = createFileRoute("/_authed/important-dates/")({
	component: DatesPage,
});

function DatesPage() {
	return (
		<div className="space-y-4">
			<h1 className="text-[18px] font-semibold tracking-tight text-ink font-display">
				Upcoming Dates
			</h1>
			<DatesList />
		</div>
	);
}
