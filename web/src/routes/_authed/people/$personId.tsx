import { createFileRoute } from "@tanstack/react-router"
import { PersonDetailSections } from "#/features/people/person-detail-sections"

export const Route = createFileRoute("/_authed/people/$personId")({
	component: PersonDetailPage,
})

function PersonDetailPage() {
	const { personId } = Route.useParams()
	return <PersonDetailSections personId={Number(personId)} />
}
