import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { createGift } from "#/endpoints/gifts"
import { keys } from "#/query-keys"
import { GiftForm } from "#/features/gifts/gift-form"
import type { GiftRequest } from "#/schemas/gift"

export const Route = createFileRoute("/_authed/gifts/new")({
	component: NewGiftPage,
})

function NewGiftPage() {
	const navigate = useNavigate()
	const qc = useQueryClient()

	const mutation = useMutation({
		mutationFn: (body: GiftRequest) => createGift(body).then(() => undefined as void),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.gifts.all })
			navigate({ to: "/gifts" })
		},
	})

	return (
		<div className="max-w-lg space-y-4">
			<h1 className="text-2xl font-heading">New Gift</h1>
			<GiftForm onSubmit={(v) => mutation.mutateAsync(v)} submitLabel="Create Gift" />
		</div>
	)
}
