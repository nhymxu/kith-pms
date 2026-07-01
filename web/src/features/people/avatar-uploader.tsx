import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Trash2, Upload } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { ImageCropDialog } from "#/components/image-crop-dialog";
import { Alert, AlertDescription } from "#/components/ui/alert";
import { Button } from "#/components/ui/button";
import { deleteAvatar, getAvatarUrl, uploadAvatar } from "#/endpoints/people";
import { keys } from "#/query-keys";

const MAX_BYTES = 5 * 1024 * 1024; // 5 MB
const ALLOWED_MIME = ["image/jpeg", "image/png", "image/gif", "image/webp"];

interface AvatarUploaderProps {
	personId: number;
	hasAvatar: boolean;
	showControls?: boolean;
}

export function AvatarUploader({
	personId,
	hasAvatar,
	showControls = true,
}: AvatarUploaderProps) {
	const inputRef = useRef<HTMLInputElement>(null);
	const [preview, setPreview] = useState<string | null>(null);
	const [clientError, setClientError] = useState<string | null>(null);
	const [cropSrc, setCropSrc] = useState<string | null>(null);
	const [cropFileName, setCropFileName] = useState("avatar");
	const qc = useQueryClient();

	const invalidate = () => {
		qc.invalidateQueries({ queryKey: keys.people.detail(personId) });
		qc.invalidateQueries({ queryKey: keys.people.avatar(personId) });
	};

	const uploadMutation = useMutation({
		mutationFn: (file: File) => uploadAvatar(personId, file),
		onSuccess: invalidate,
	});

	const deleteMutation = useMutation({
		mutationFn: () => deleteAvatar(personId),
		onSuccess: () => {
			setPreview((prev) => {
				if (prev) URL.revokeObjectURL(prev);
				return null;
			});
			invalidate();
		},
	});

	const previewRef = useRef<string | null>(null);
	previewRef.current = preview;
	useEffect(() => {
		return () => {
			if (previewRef.current) URL.revokeObjectURL(previewRef.current);
		};
	}, []);

	function handleFile(file: File) {
		if (cropSrc) return; // crop dialog already open for a previous pick
		setClientError(null);
		if (!ALLOWED_MIME.includes(file.type)) {
			setClientError("Only JPEG, PNG, GIF, or WebP images are allowed.");
			return;
		}
		if (file.size > MAX_BYTES) {
			setClientError("File must be under 5 MB.");
			return;
		}
		setCropFileName(file.name);
		setCropSrc(URL.createObjectURL(file));
	}

	function handleCropCancel() {
		if (cropSrc) URL.revokeObjectURL(cropSrc);
		setCropSrc(null);
	}

	function handleCropped(file: File) {
		if (cropSrc) URL.revokeObjectURL(cropSrc);
		setCropSrc(null);
		if (file.size > MAX_BYTES) {
			setClientError("Cropped image is too large — try a smaller zoom area.");
			return;
		}
		setPreview((prev) => {
			if (prev) URL.revokeObjectURL(prev);
			return URL.createObjectURL(file);
		});
		uploadMutation.mutate(file);
	}

	function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
		const file = e.target.files?.[0];
		if (file) handleFile(file);
		// Reset so the same file can be re-selected
		e.target.value = "";
	}

	function handleDrop(e: React.DragEvent<HTMLButtonElement>) {
		e.preventDefault();
		const file = e.dataTransfer.files?.[0];
		if (file) handleFile(file);
	}

	const currentSrc = preview ?? (hasAvatar ? getAvatarUrl(personId) : null);
	const isPending = uploadMutation.isPending || deleteMutation.isPending;

	return (
		<div className="space-y-3">
			{/* Drop zone */}
			<button
				type="button"
				tabIndex={showControls ? 0 : -1}
				onDrop={handleDrop}
				onDragOver={(e) => e.preventDefault()}
				onClick={() => showControls && inputRef.current?.click()}
				onKeyDown={(e) =>
					e.key === "Enter" && showControls && inputRef.current?.click()
				}
				className={`w-24 h-24 rounded-md border border-dashed border-zinc-300 overflow-hidden bg-secondary-background flex items-center justify-center transition-colors ${showControls ? "cursor-pointer hover:border-main" : "cursor-default"}`}
			>
				{currentSrc ? (
					<img
						src={currentSrc}
						alt="Avatar"
						className="size-full object-cover"
					/>
				) : (
					<Upload className="size-6 text-foreground/40" />
				)}
			</button>

			<input
				ref={inputRef}
				type="file"
				accept={ALLOWED_MIME.join(",")}
				className="hidden"
				onChange={handleChange}
			/>

			{showControls && (
				<div className="flex gap-2">
					<Button
						type="button"
						variant="neutral"
						size="sm"
						disabled={isPending}
						onClick={() => inputRef.current?.click()}
					>
						<Upload className="size-3" />
						{hasAvatar || preview ? "Replace" : "Upload"}
					</Button>
					{(hasAvatar || preview) && (
						<Button
							type="button"
							variant="destructive"
							size="sm"
							disabled={isPending}
							onClick={() => deleteMutation.mutate()}
						>
							<Trash2 className="size-3" /> Remove
						</Button>
					)}
				</div>
			)}

			{(clientError ?? uploadMutation.error) && (
				<Alert variant="destructive">
					<AlertDescription>
						{clientError ??
							(uploadMutation.error instanceof Error
								? uploadMutation.error.message
								: "Upload failed")}
					</AlertDescription>
				</Alert>
			)}

			{cropSrc && (
				<ImageCropDialog
					open
					imageSrc={cropSrc}
					fileName={cropFileName}
					aspect={1}
					onCancel={handleCropCancel}
					onCropped={handleCropped}
				/>
			)}
		</div>
	);
}
