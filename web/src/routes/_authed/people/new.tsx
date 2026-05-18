import { createFileRoute } from "@tanstack/react-router"
import { PersonForm } from "#/features/people/person-form"

export const Route = createFileRoute("/_authed/people/new")({
	component: NewPersonPage,
})

function NewPersonPage() {
	return (
		<div className="space-y-4">
			<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">New Person</h1>
			<PersonForm mode="create" />
		</div>
	)
}
