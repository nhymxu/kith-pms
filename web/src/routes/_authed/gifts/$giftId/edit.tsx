import {
	useMutation,
	useQueryClient,
	useSuspenseQuery,
} from "@tanstack/react-query";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { Button } from "#/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "#/components/ui/card";
import {
	deleteGiftImage,
	getGift,
	updateGift,
	uploadGiftImage,
} from "#/endpoints/gifts";
import { GiftForm } from "#/features/gifts/gift-form";
import { GiftImageUploader } from "#/features/gifts/gift-image-uploader";
import { keys } from "#/query-keys";
import type { GiftRequest } from "#/schemas/gift";

export const Route = createFileRoute("/_authed/gifts/$giftId/edit")({
	component: EditGiftPage,
	pendingComponent: () => (
		<p className="text-sm font-base text-foreground/60">Loading…</p>
	),
	errorComponent: () => (
		<p className="text-sm font-base text-destructive">Gift not found.</p>
	),
});

function EditGiftPage() {
	const { giftId } = Route.useParams();
	const id = Number(giftId);
	const navigate = useNavigate();
	const qc = useQueryClient();

	const { data } = useSuspenseQuery({
		queryKey: keys.gifts.detail(id),
		queryFn: () => getGift(id),
	});

	const mutation = useMutation({
		mutationFn: (body: GiftRequest) => updateGift(id, body),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.gifts.detail(id) });
			qc.invalidateQueries({ queryKey: keys.gifts.all });
			navigate({ to: "/gifts/$giftId", params: { giftId } });
		},
	});

	const uploadImageMutation = useMutation({
		mutationFn: (file: File) => uploadGiftImage(id, file),
		onSuccess: () => qc.invalidateQueries({ queryKey: keys.gifts.detail(id) }),
	});

	const removeImageMutation = useMutation({
		mutationFn: () => deleteGiftImage(id),
		onSuccess: () => qc.invalidateQueries({ queryKey: keys.gifts.detail(id) }),
	});

	const handleCancel = () =>
		navigate({ to: "/gifts/$giftId", params: { giftId } });

	return (
		<div className="max-w-lg space-y-4">
			<div className="flex items-center justify-between">
				<div>
					<p className="text-xs text-sub uppercase tracking-wide font-medium">
						Editing
					</p>
					<h1 className="text-[18px] font-semibold tracking-tight text-ink font-display">
						{data.title}
					</h1>
				</div>
				<Button variant="neutral" onClick={handleCancel}>
					Cancel
				</Button>
			</div>
			<GiftForm
				initial={data}
				onSubmit={(v) => mutation.mutateAsync(v)}
				submitLabel="Save changes"
				onCancel={handleCancel}
			/>

			<Card>
				<CardHeader>
					<CardTitle className="text-base">Image</CardTitle>
				</CardHeader>
				<CardContent>
					<GiftImageUploader
						imageSrc={data.image_path ? `/v1/gifts/${id}/image` : null}
						hasImage={Boolean(data.image_path)}
						onUpload={(file) => uploadImageMutation.mutate(file)}
						onRemove={() => removeImageMutation.mutate()}
						uploadPending={uploadImageMutation.isPending}
						uploadError={uploadImageMutation.isError ? "Upload failed." : null}
						isRemoving={removeImageMutation.isPending}
					/>
				</CardContent>
			</Card>
		</div>
	);
}
