import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { getGift, updateGift } from "#/endpoints/gifts"
import { keys } from "#/query-keys"
import { GiftForm } from "#/features/gifts/gift-form"
import type { GiftRequest } from "#/schemas/gift"

export const Route = createFileRoute("/_authed/gifts/$giftId/edit")({
	component: EditGiftPage,
})

function EditGiftPage() {
	const { giftId } = Route.useParams()
	const id = Number(giftId)
	const navigate = useNavigate()
	const qc = useQueryClient()

	const { data, isPending, isError } = useQuery({
		queryKey: keys.gifts.detail(id),
		queryFn: () => getGift(id),
	})

	const mutation = useMutation({
		mutationFn: (body: GiftRequest) => updateGift(id, body),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.gifts.detail(id) })
			qc.invalidateQueries({ queryKey: keys.gifts.all })
			navigate({ to: "/gifts/$giftId", params: { giftId } })
		},
	})

	if (isPending) return <p className="text-sm font-base text-foreground/60">Loading…</p>
	if (isError || !data) return <p className="text-sm font-base text-destructive">Gift not found.</p>

	return (
		<div className="max-w-lg space-y-4">
			<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">Edit Gift</h1>
			<GiftForm
				initial={data}
				onSubmit={(v) => mutation.mutateAsync(v)}
				submitLabel="Update Gift"
			/>
		</div>
	)
}
