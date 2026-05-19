import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { getGift, updateGift, uploadGiftImage, deleteGiftImage } from "#/endpoints/gifts"
import { keys } from "#/query-keys"
import { GiftForm } from "#/features/gifts/gift-form"
import { Button } from "#/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "#/components/ui/card"
import { Label } from "#/components/ui/label"
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

	const uploadImageMutation = useMutation({
		mutationFn: (file: File) => uploadGiftImage(id, file),
		onSuccess: () => qc.invalidateQueries({ queryKey: keys.gifts.detail(id) }),
	})

	const removeImageMutation = useMutation({
		mutationFn: () => deleteGiftImage(id),
		onSuccess: () => qc.invalidateQueries({ queryKey: keys.gifts.detail(id) }),
	})

	if (isPending) return <p className="text-sm font-base text-foreground/60">Loading…</p>
	if (isError || !data) return <p className="text-sm font-base text-destructive">Gift not found.</p>

	const handleCancel = () => navigate({ to: "/gifts/$giftId", params: { giftId } })

	return (
		<div className="max-w-lg space-y-4">
			<div className="flex items-center justify-between">
				<div>
					<p className="text-xs text-zinc-400 uppercase tracking-wide font-medium">Editing</p>
					<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">{data.title}</h1>
				</div>
				<Button variant="neutral" onClick={handleCancel}>Cancel</Button>
			</div>
			<GiftForm
				initial={data}
				onSubmit={(v) => mutation.mutateAsync(v)}
				submitLabel="Save changes"
				onCancel={handleCancel}
			/>

			<Card>
				<CardHeader><CardTitle className="text-base">Image</CardTitle></CardHeader>
				<CardContent className="space-y-3">
					{data.image_path ? (
						<div className="space-y-2">
							<img
								src={`/v1/gifts/${id}/image`}
								alt={data.title}
								className="max-h-48 rounded border border-zinc-200 object-contain"
							/>
							<Button
								variant="destructive"
								size="sm"
								onClick={() => removeImageMutation.mutate()}
								disabled={removeImageMutation.isPending}
							>
								{removeImageMutation.isPending ? "Removing…" : "Remove image"}
							</Button>
						</div>
					) : (
						<p className="text-sm text-zinc-500">No image uploaded.</p>
					)}
					<div className="space-y-1.5">
						<Label>Upload new image</Label>
						<input
							type="file"
							accept="image/jpeg,image/png,image/gif,image/webp"
							onChange={(e) => {
								const f = e.target.files?.[0]
								if (f) uploadImageMutation.mutate(f)
							}}
							className="text-sm"
						/>
						{uploadImageMutation.isPending && <p className="text-xs text-zinc-500">Uploading…</p>}
						{uploadImageMutation.isError && <p className="text-xs text-red-600">Upload failed.</p>}
					</div>
				</CardContent>
			</Card>
		</div>
	)
}
