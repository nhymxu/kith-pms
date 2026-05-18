import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { z } from "zod"
import { createJournalEntry } from "#/endpoints/journal"
import { keys } from "#/query-keys"
import { JournalForm } from "#/features/journal/journal-form"
import type { JournalRequest } from "#/schemas/journal"

const searchSchema = z.object({
	person_id: z.coerce.number().optional(),
})

export const Route = createFileRoute("/_authed/journal/new")({
	validateSearch: searchSchema,
	component: NewJournalPage,
})

function NewJournalPage() {
	const navigate = useNavigate()
	const qc = useQueryClient()
	const { person_id } = Route.useSearch()

	const mutation = useMutation({
		mutationFn: async (body: JournalRequest) => {
			const id = await createJournalEntry(body)
			qc.invalidateQueries({ queryKey: keys.journal.all })
			navigate({ to: "/journal/$entryId", params: { entryId: String(id) } })
		},
	})

	return (
		<div className="space-y-4">
			<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">New Journal Entry</h1>
			<JournalForm
				onSubmit={(v) => mutation.mutateAsync(v)}
				submitLabel="Create Entry"
				defaultPersonId={person_id}
			/>
		</div>
	)
}
