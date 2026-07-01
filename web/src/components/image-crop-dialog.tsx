import { useState } from "react";
import Cropper, { type Area } from "react-easy-crop";
import { Button } from "#/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import { Slider } from "#/components/ui/slider";
import { blobToFile, cropImageToBlob } from "#/lib/crop-image";

interface ImageCropDialogProps {
	open: boolean;
	imageSrc: string;
	fileName: string;
	/** Fixed crop viewport aspect ratio (width / height). */
	aspect: number;
	onCancel: () => void;
	onCropped: (file: File) => void;
}

export function ImageCropDialog({
	open,
	imageSrc,
	fileName,
	aspect,
	onCancel,
	onCropped,
}: ImageCropDialogProps) {
	const [crop, setCrop] = useState({ x: 0, y: 0 });
	const [zoom, setZoom] = useState(1);
	const [cropArea, setCropArea] = useState<Area | null>(null);
	const [isSaving, setIsSaving] = useState(false);
	const [error, setError] = useState<string | null>(null);

	async function handleSave() {
		if (!cropArea) return;
		setIsSaving(true);
		setError(null);
		try {
			const blob = await cropImageToBlob(imageSrc, cropArea);
			onCropped(blobToFile(blob, fileName));
		} catch {
			setError("Couldn't crop this image. Try a different file.");
		} finally {
			setIsSaving(false);
		}
	}

	return (
		<Dialog open={open} onOpenChange={(next) => !next && onCancel()}>
			<DialogContent className="max-w-[calc(100%-2rem)] sm:max-w-2xl">
				<DialogHeader>
					<DialogTitle>Crop image</DialogTitle>
				</DialogHeader>

				<div className="relative h-[min(70vh,32rem)] bg-secondary-background rounded-md overflow-hidden">
					<Cropper
						image={imageSrc}
						crop={crop}
						zoom={zoom}
						aspect={aspect}
						onCropChange={setCrop}
						onZoomChange={setZoom}
						onCropComplete={(_, areaPixels) => setCropArea(areaPixels)}
					/>
				</div>

				<div className="space-y-1.5">
					<span className="text-xs text-foreground/60">Zoom</span>
					<Slider
						min={1}
						max={3}
						step={0.01}
						value={[zoom]}
						onValueChange={(v) => setZoom(Array.isArray(v) ? (v[0] ?? 1) : v)}
					/>
				</div>

				{error && <p className="text-xs text-destructive">{error}</p>}

				<DialogFooter>
					<Button type="button" variant="neutral" onClick={onCancel}>
						Cancel
					</Button>
					<Button type="button" onClick={handleSave} disabled={isSaving}>
						{isSaving ? "Cropping…" : "Save crop"}
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
