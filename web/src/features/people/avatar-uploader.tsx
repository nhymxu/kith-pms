import {
	useMutation,
	useQueryClient,
	useSuspenseQuery,
} from "@tanstack/react-query";
import { Trash2, Upload } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { ImageCropDialog } from "#/components/image-crop-dialog";
import { Alert, AlertDescription } from "#/components/ui/alert";
import { Button } from "#/components/ui/button";
import { deleteAvatar, getAvatarUrl, uploadAvatar } from "#/endpoints/people";
import { getSettings } from "#/endpoints/settings";
import { useImageFilePick } from "#/hooks/use-image-file-pick";
import { FILE_INPUT_ACCEPT } from "#/lib/image-constraints";
import { keys } from "#/query-keys";

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

	const { data: settings } = useSuspenseQuery({
		queryKey: ["settings"],
		queryFn: getSettings,
	});
	const maxBytes = settings.max_upload_size_mb * 1024 * 1024;

	const {
		prepare,
		isConverting,
		error: pickError,
		setError: setPickError,
	} = useImageFilePick({
		maxBytes,
		maxSizeMB: settings.max_upload_size_mb,
	});

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

	// Unmounting with a crop dialog open (navigating away without cancelling)
	// would otherwise leak that object URL until a page reload.
	useEffect(() => {
		return () => {
			if (previewRef.current) URL.revokeObjectURL(previewRef.current);
			if (cropSrc) URL.revokeObjectURL(cropSrc);
		};
	}, [cropSrc]);

	async function handleFile(file: File) {
		if (cropSrc || isConverting) return; // a pick is already in flight
		setClientError(null);
		setPickError(null);

		const prepared = await prepare(file);
		if (!prepared) return;

		setCropFileName(prepared.name);
		setCropSrc(URL.createObjectURL(prepared));
	}

	function handleCropCancel() {
		if (cropSrc) URL.revokeObjectURL(cropSrc);
		setCropSrc(null);
	}

	function handleCropped(file: File) {
		if (cropSrc) URL.revokeObjectURL(cropSrc);
		setCropSrc(null);
		setPreview((prev) => {
			if (prev) URL.revokeObjectURL(prev);
			return URL.createObjectURL(file);
		});
		uploadMutation.mutate(file);
	}

	function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
		const file = e.target.files?.[0];
		if (file) void handleFile(file);
		// Reset so the same file can be re-selected
		e.target.value = "";
	}

	function handleDrop(e: React.DragEvent<HTMLButtonElement>) {
		e.preventDefault();
		const file = e.dataTransfer.files?.[0];
		if (file) void handleFile(file);
	}

	const currentSrc = preview ?? (hasAvatar ? getAvatarUrl(personId) : null);
	const isPending =
		uploadMutation.isPending || deleteMutation.isPending || isConverting;

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
				className={`w-24 h-24 rounded-md border-bw border-dashed border-line overflow-hidden bg-chip flex items-center justify-center transition-colors ${showControls ? "cursor-pointer hover:border-accent" : "cursor-default"}`}
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
				accept={FILE_INPUT_ACCEPT}
				className="hidden"
				disabled={isPending}
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
						{isConverting
							? "Converting…"
							: hasAvatar || preview
								? "Replace"
								: "Upload"}
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

			{(pickError ?? clientError ?? uploadMutation.error) && (
				<Alert variant="destructive">
					<AlertDescription>
						{pickError ??
							clientError ??
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
					aspectRatio={1}
					maxEdgePx={settings.image_max_edge_px}
					quality={settings.image_jpeg_quality}
					onCancel={handleCropCancel}
					onCropped={handleCropped}
				/>
			)}
		</div>
	);
}
