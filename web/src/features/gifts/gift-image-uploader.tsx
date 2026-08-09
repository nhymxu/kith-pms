// Shared gift-image upload surface: a bordered drop-zone box + explicit
// Upload/Replace and Remove buttons, mirroring `avatar-uploader.tsx`.
import { useSuspenseQuery } from "@tanstack/react-query";
import { Trash2, Upload } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { ImageCropDialog } from "#/components/image-crop-dialog";
import { Button } from "#/components/ui/button";
import { getSettings } from "#/endpoints/settings";
import { useImageFilePick } from "#/hooks/use-image-file-pick";
import { FILE_INPUT_ACCEPT } from "#/lib/image-constraints";

interface GiftImageUploaderProps {
	/** Current image to show in the drop zone (edit: server URL; new: crop preview). */
	imageSrc?: string | null;
	/** True when an image is currently set, controlling Remove / Replace labels. */
	hasImage: boolean;
	/** Uploads the cropped result. */
	onUpload: (file: File) => void;
	/** Clears the current image (edit: DELETE endpoint; new: clear pending preview). */
	onRemove: () => void;
	/** Upload mutation progress. */
	uploadPending: boolean;
	/** Upload mutation failure message, if any. */
	uploadError?: string | null;
	/** Remove mutation in progress, disabling Remove. */
	isRemoving?: boolean;
}

export function GiftImageUploader({
	imageSrc,
	hasImage,
	onUpload,
	onRemove,
	uploadPending,
	uploadError,
	isRemoving = false,
}: GiftImageUploaderProps) {
	const inputRef = useRef<HTMLInputElement>(null);
	const [cropSrc, setCropSrc] = useState<string | null>(null);
	const cropSrcRef = useRef<string | null>(null);
	cropSrcRef.current = cropSrc;
	const [cropFileName, setCropFileName] = useState("gift-image");
	const [imageError, setImageError] = useState<string | null>(null);
	// Optimistic preview of the freshly cropped pick so the box shows the new
	// image immediately, before the server round-trip — mirrors avatar-uploader.
	const [localPreview, setLocalPreview] = useState<string | null>(null);
	const localPreviewRef = useRef<string | null>(null);
	localPreviewRef.current = localPreview;

	// Navigating away with the crop dialog open would otherwise leak the picked
	// image's object URL until a reload — same cleanup avatar-uploader does.
	useEffect(() => {
		return () => {
			if (cropSrcRef.current) URL.revokeObjectURL(cropSrcRef.current);
			if (localPreviewRef.current) URL.revokeObjectURL(localPreviewRef.current);
		};
	}, []);

	// Route matching may reuse this component across edits, so reset optimistic
	// crop state when the underlying image identity changes (new gift id, or a
	// previously-empty image that is now set).
	useEffect(() => {
		void imageSrc; // identity signal only — the reset below is value-agnostic
		setCropSrc((prev) => {
			if (prev) URL.revokeObjectURL(prev);
			return null;
		});
		setLocalPreview((prev) => {
			if (prev) URL.revokeObjectURL(prev);
			return null;
		});
	}, [imageSrc]);

	// Roll the box back to the caller-provided image when the upload fails, so
	// it stops showing a crop the server never accepted (uploadError flips non-null).
	useEffect(() => {
		if (uploadError) {
			setLocalPreview((prev) => {
				if (prev) URL.revokeObjectURL(prev);
				return null;
			});
		}
	}, [uploadError]);

	const { data: settings } = useSuspenseQuery({
		queryKey: ["settings"],
		queryFn: getSettings,
	});
	const maxImageBytes = settings.max_upload_size_mb * 1024 * 1024;
	const controlsDisabled = uploadPending || isRemoving;

	const {
		prepare: prepareImage,
		isConverting,
		error: pickError,
		setError: setPickError,
	} = useImageFilePick({
		maxBytes: maxImageBytes,
		maxSizeMB: settings.max_upload_size_mb,
	});

	async function handleFile(file: File) {
		if (cropSrc || isConverting) return; // a pick is already in flight
		setImageError(null);
		setPickError(null);
		const prepared = await prepareImage(file);
		if (!prepared) return;
		setCropFileName(prepared.name);
		setCropSrc(URL.createObjectURL(prepared));
	}

	function handleChangeEvent(e: React.ChangeEvent<HTMLInputElement>) {
		const file = e.target.files?.[0];
		e.target.value = "";
		if (file) void handleFile(file);
	}

	function handleDrop(e: React.DragEvent<HTMLButtonElement>) {
		e.preventDefault();
		const file = e.dataTransfer.files?.[0];
		if (file) void handleFile(file);
	}

	function handleCropCancel() {
		if (cropSrc) URL.revokeObjectURL(cropSrc);
		setCropSrc(null);
	}

	function handleCropped(file: File) {
		if (cropSrc) URL.revokeObjectURL(cropSrc);
		setCropSrc(null);
		if (file.size > maxImageBytes) {
			setImageError("Cropped image is too large — try a smaller crop area.");
			return;
		}
		// Show the baked crop immediately; the caller's imageSrc (edit: server
		// URL) only reflects it after the upload round-trip.
		const nextPreview = URL.createObjectURL(file);
		setLocalPreview((prev) => {
			if (prev) URL.revokeObjectURL(prev);
			return nextPreview;
		});
		onUpload(file);
	}

	function handleRemove() {
		// Drop the optimistic preview so the box reflects the removal even
		// before the caller's imageSrc (edit: refetch) updates.
		setLocalPreview((prev) => {
			if (prev) URL.revokeObjectURL(prev);
			return null;
		});
		onRemove();
	}

	const openPicker = () => inputRef.current?.click();
	// The optimistic crop preview wins over the caller-provided URL so a fresh
	// crop (or its removal) is always reflected without waiting on the server.
	const displaySrc = localPreview ?? imageSrc;
	const message = pickError ?? imageError ?? uploadError;

	return (
		<div className="space-y-3">
			{/* Drop zone / current image preview */}
			<button
				type="button"
				tabIndex={0}
				onDrop={handleDrop}
				onDragOver={(e) => e.preventDefault()}
				onClick={openPicker}
				onKeyDown={(e) => {
					if (e.key !== "Enter") return;
					e.preventDefault(); // avoid the duplicate activation click
					openPicker();
				}}
				aria-label={hasImage ? "Replace gift image" : "Upload gift image"}
				disabled={controlsDisabled || isConverting}
				className="h-32 w-32 rounded-md border-bw border-dashed border-line overflow-hidden bg-chip flex items-center justify-center transition-colors cursor-pointer hover:border-accent disabled:cursor-not-allowed"
			>
				{displaySrc ? (
					<img src={displaySrc} alt="" className="size-full object-cover" />
				) : (
					<Upload className="size-6 text-sub" />
				)}
			</button>

			<input
				ref={inputRef}
				type="file"
				accept={FILE_INPUT_ACCEPT}
				className="hidden"
				disabled={isConverting}
				onChange={handleChangeEvent}
			/>

			<div className="flex gap-2">
				<Button
					type="button"
					variant="neutral"
					size="sm"
					disabled={controlsDisabled}
					onClick={openPicker}
				>
					<Upload className="size-3" />
					{isConverting ? "Converting…" : hasImage ? "Replace" : "Upload"}
				</Button>
				{hasImage && (
					<Button
						type="button"
						variant="destructive"
						size="sm"
						disabled={controlsDisabled}
						onClick={handleRemove}
					>
						<Trash2 className="size-3" /> {isRemoving ? "Removing…" : "Remove"}
					</Button>
				)}
			</div>

			{uploadPending && <p className="text-xs text-sub">Uploading image…</p>}
			{message && <p className="text-xs text-danger-fg">{message}</p>}

			{cropSrc && (
				<ImageCropDialog
					open
					imageSrc={cropSrc}
					fileName={cropFileName}
					maxEdgePx={settings.image_max_edge_px}
					quality={settings.image_jpeg_quality}
					onCancel={handleCropCancel}
					onCropped={handleCropped}
				/>
			)}
		</div>
	);
}
