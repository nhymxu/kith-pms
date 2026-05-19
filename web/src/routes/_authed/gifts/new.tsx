import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useRef } from "react"
import { createGift, uploadGiftImage } from "#/endpoints/gifts"
import { keys } from "#/query-keys"
import { GiftForm } from "#/features/gifts/gift-form"
import type { GiftRequest } from "#/schemas/gift"

export const Route = createFileRoute("/_authed/gifts/new")({
	component: NewGiftPage,
})

function NewGiftPage() {
	const navigate = useNavigate()
	const qc = useQueryClient()
	const pendingImage = useRef<File | null>(null)

	const mutation = useMutation({
		mutationFn: async (body: GiftRequest): Promise<void> => {
			const id = await createGift(body)
			if (pendingImage.current) {
				await uploadGiftImage(id, pendingImage.current)
			}
		},
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.gifts.all })
			navigate({ to: "/gifts" })
		},
	})

	return (
		<div className="max-w-lg space-y-4">
			<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">New Gift</h1>
			<GiftForm
				onSubmit={(v) => mutation.mutateAsync(v)}
				submitLabel="Create Gift"
				onImageChange={(f) => { pendingImage.current = f }}
			/>
		</div>
	)
}
