import {
	useMutation,
	useQueryClient,
	useSuspenseQuery,
} from "@tanstack/react-query";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { ImageCropDialog } from "#/components/image-crop-dialog";
import { Button } from "#/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "#/components/ui/card";
import { Label } from "#/components/ui/label";
import {
	deleteGiftImage,
	getGift,
	updateGift,
	uploadGiftImage,
} from "#/endpoints/gifts";
import { getSettings } from "#/endpoints/settings";
import { GiftForm } from "#/features/gifts/gift-form";
import { GIFT_IMAGE_ASPECT } from "#/features/gifts/gift-image-constraints";
import { useImageFilePick } from "#/hooks/use-image-file-pick";
import { FILE_INPUT_ACCEPT } from "#/lib/image-constraints";
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
	const [cropSrc, setCropSrc] = useState<string | null>(null);
	const [cropFileName, setCropFileName] = useState("gift-image");
	const [imageError, setImageError] = useState<string | null>(null);

	const { data } = useSuspenseQuery({
		queryKey: keys.gifts.detail(id),
		queryFn: () => getGift(id),
	});

	const { data: settings } = useSuspenseQuery({
		queryKey: ["settings"],
		queryFn: getSettings,
	});
	const maxImageBytes = settings.max_upload_size_mb * 1024 * 1024;

	const {
		prepare: prepareImage,
		isConverting,
		error: pickError,
		setError: setPickError,
	} = useImageFilePick({
		maxBytes: maxImageBytes,
		maxSizeMB: settings.max_upload_size_mb,
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
				<CardContent className="space-y-3">
					{data.image_path ? (
						<div className="space-y-2">
							<img
								src={`/v1/gifts/${id}/image`}
								alt={data.title}
								className="max-h-48 rounded-base border-bw border-line object-contain"
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
						<p className="text-sm text-sub">No image uploaded.</p>
					)}
					<div className="space-y-1.5">
						<Label>Upload new image</Label>
						<input
							type="file"
							accept={FILE_INPUT_ACCEPT}
							disabled={isConverting}
							onChange={(e) => {
								const file = e.target.files?.[0];
								e.target.value = "";
								if (!file || cropSrc || isConverting) return;
								setImageError(null);
								setPickError(null);
								void prepareImage(file)
									.then((prepared) => {
										if (!prepared) return;
										setCropFileName(prepared.name);
										setCropSrc(URL.createObjectURL(prepared));
									})
									.catch(() => setImageError("Couldn't read this image."));
							}}
							className="text-sm"
						/>
						{isConverting && (
							<p className="text-xs text-sub">Converting image…</p>
						)}
						{(pickError ?? imageError) && (
							<p className="text-xs text-danger-fg">
								{pickError ?? imageError}
							</p>
						)}
						{uploadImageMutation.isPending && (
							<p className="text-xs text-sub">Uploading…</p>
						)}
						{uploadImageMutation.isError && (
							<p className="text-xs text-danger-fg">Upload failed.</p>
						)}
					</div>
				</CardContent>
			</Card>

			{cropSrc && (
				<ImageCropDialog
					open
					imageSrc={cropSrc}
					fileName={cropFileName}
					aspect={GIFT_IMAGE_ASPECT}
					maxEdgePx={settings.image_max_edge_px}
					quality={settings.image_jpeg_quality}
					onCancel={() => {
						URL.revokeObjectURL(cropSrc);
						setCropSrc(null);
					}}
					onCropped={(file) => {
						URL.revokeObjectURL(cropSrc);
						setCropSrc(null);
						if (file.size > maxImageBytes) {
							setImageError(
								"Cropped image is too large — try a smaller zoom area.",
							);
							return;
						}
						uploadImageMutation.mutate(file);
					}}
				/>
			)}
		</div>
	);
}
