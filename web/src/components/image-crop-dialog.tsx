import { useRef, useState } from "react";
import {
	Cropper,
	type CropperRef,
	RectangleStencil,
} from "react-advanced-cropper";
import "react-advanced-cropper/dist/style.css";
import { Button } from "#/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import {
	blobToFile,
	cropImageToBlob,
	encodeCropped,
	loadImageWithSize,
} from "#/lib/crop-image";

interface ImageCropDialogProps {
	open: boolean;
	imageSrc: string;
	fileName: string;
	/**
	 * Crop-box aspect ratio (width / height). Omitted (gift images) gives a
	 * free-form, user-resizable rectangle with an optional skip; provided (avatar
	 * passes 1) pins the box and makes a crop mandatory.
	 */
	aspectRatio?: number;
	/** Longest edge of the encoded result; from server config. */
	maxEdgePx: number;
	/** JPEG quality 1-100; from server config. */
	quality: number;
	onCancel: () => void;
	onCropped: (file: File) => void;
}

export function ImageCropDialog({
	open,
	imageSrc,
	fileName,
	aspectRatio,
	maxEdgePx,
	quality,
	onCancel,
	onCropped,
}: ImageCropDialogProps) {
	const [isSaving, setIsSaving] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const cropperRef = useRef<CropperRef | null>(null);

	/** Free-form mode: the user picks the crop and may skip it entirely. */
	const freeform = aspectRatio == null;

	/** Runs `action` with the saving flag, mapping failures to a generic message. */
	async function withProcessing(action: () => Promise<void>) {
		setIsSaving(true);
		setError(null);
		try {
			await action();
		} catch {
			setError("Couldn't process this image. Try a different file.");
		} finally {
			setIsSaving(false);
		}
	}

	async function handleSave() {
		const coords = cropperRef.current?.getCoordinates();
		if (!coords) {
			setError("Couldn't crop this image. Try a different file.");
			return;
		}
		await withProcessing(async () => {
			const blob = await cropImageToBlob(
				imageSrc,
				{
					x: coords.left,
					y: coords.top,
					width: coords.width,
					height: coords.height,
				},
				{ maxEdgePx, quality },
			);
			onCropped(blobToFile(blob, fileName));
		});
	}

	/** Skip cropping: encode the full image, reusing the single decode. */
	async function handleSkip() {
		await withProcessing(async () => {
			const { image, width, height } = await loadImageWithSize(imageSrc);
			const blob = await encodeCropped(
				image,
				{ x: 0, y: 0, width, height },
				{ maxEdgePx, quality },
			);
			onCropped(blobToFile(blob, fileName));
		});
	}

	return (
		<Dialog open={open} onOpenChange={(next) => !next && onCancel()}>
			<DialogContent className="max-w-[calc(100%-2rem)] sm:max-w-2xl">
				<DialogHeader>
					<DialogTitle>{freeform ? "Adjust image" : "Crop image"}</DialogTitle>
				</DialogHeader>

				<div className="relative h-[min(70vh,32rem)] bg-chip rounded-md overflow-hidden">
					<Cropper
						ref={cropperRef}
						src={imageSrc}
						className="h-full"
						stencilComponent={RectangleStencil}
						stencilProps={
							freeform
								? {
										minAspectRatio: 1 / 4,
										handlers: {
											north: false,
											south: false,
											east: false,
											west: false,
										},
									}
								: { aspectRatio }
						}
					/>
				</div>

				{freeform && (
					<p className="text-xs text-foreground/60">
						Drag the corners to choose the crop area, or skip cropping to keep
						the whole image.
					</p>
				)}

				{error && <p className="text-xs text-destructive">{error}</p>}

				<DialogFooter>
					<Button type="button" variant="neutral" onClick={onCancel}>
						Cancel
					</Button>
					{freeform && (
						<Button
							type="button"
							variant="neutral"
							onClick={handleSkip}
							disabled={isSaving}
						>
							Skip crop
						</Button>
					)}
					<Button type="button" onClick={handleSave} disabled={isSaving}>
						{isSaving ? "Processing…" : freeform ? "Apply" : "Save crop"}
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
