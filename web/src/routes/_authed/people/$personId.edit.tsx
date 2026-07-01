import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { getPerson } from "#/endpoints/people";
import { PersonForm } from "#/features/people/person-form";
import { keys } from "#/query-keys";

export const Route = createFileRoute("/_authed/people/$personId/edit")({
	component: EditPersonPage,
	pendingComponent: () => (
		<p className="text-sm font-base text-foreground/60">Loading…</p>
	),
	errorComponent: () => (
		<p className="text-sm font-base text-destructive">Person not found.</p>
	),
});

function EditPersonPage() {
	const { personId } = Route.useParams();
	const id = Number(personId);

	const { data } = useSuspenseQuery({
		queryKey: keys.people.detail(id),
		queryFn: () => getPerson(id),
	});

	return (
		<div className="space-y-4">
			<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">
				Edit {data.name}
			</h1>
			<PersonForm mode="edit" initial={data} />
		</div>
	);
}
