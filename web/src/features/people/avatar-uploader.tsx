import { useRef, useState } from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { Upload, Trash2 } from "lucide-react"
import { Button } from "#/components/ui/button"
import { Alert, AlertDescription } from "#/components/ui/alert"
import { keys } from "#/query-keys"
import { uploadAvatar, deleteAvatar, getAvatarUrl } from "#/endpoints/people"

const MAX_BYTES = 5 * 1024 * 1024 // 5 MB
const ALLOWED_MIME = ["image/jpeg", "image/png", "image/gif", "image/webp"]

interface AvatarUploaderProps {
	personId: number
	hasAvatar: boolean
}

export function AvatarUploader({ personId, hasAvatar }: AvatarUploaderProps) {
	const inputRef = useRef<HTMLInputElement>(null)
	const [preview, setPreview] = useState<string | null>(null)
	const [clientError, setClientError] = useState<string | null>(null)
	const qc = useQueryClient()

	const invalidate = () => {
		qc.invalidateQueries({ queryKey: keys.people.detail(personId) })
		qc.invalidateQueries({ queryKey: keys.people.avatar(personId) })
	}

	const uploadMutation = useMutation({
		mutationFn: (file: File) => uploadAvatar(personId, file),
		onSuccess: invalidate,
	})

	const deleteMutation = useMutation({
		mutationFn: () => deleteAvatar(personId),
		onSuccess: () => { setPreview(null); invalidate() },
	})

	function handleFile(file: File) {
		setClientError(null)
		if (!ALLOWED_MIME.includes(file.type)) {
			setClientError("Only JPEG, PNG, GIF, or WebP images are allowed.")
			return
		}
		if (file.size > MAX_BYTES) {
			setClientError("File must be under 5 MB.")
			return
		}
		const url = URL.createObjectURL(file)
		setPreview(url)
		uploadMutation.mutate(file)
	}

	function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
		const file = e.target.files?.[0]
		if (file) handleFile(file)
		// Reset so the same file can be re-selected
		e.target.value = ""
	}

	function handleDrop(e: React.DragEvent<HTMLDivElement>) {
		e.preventDefault()
		const file = e.dataTransfer.files?.[0]
		if (file) handleFile(file)
	}

	const currentSrc = preview ?? (hasAvatar ? getAvatarUrl(personId) : null)
	const isPending = uploadMutation.isPending || deleteMutation.isPending

	return (
		<div className="space-y-3">
			{/* Drop zone */}
			<div
				onDrop={handleDrop}
				onDragOver={(e) => e.preventDefault()}
				onClick={() => inputRef.current?.click()}
				className="w-24 h-24 rounded-base border-2 border-dashed border-border cursor-pointer overflow-hidden bg-secondary-background flex items-center justify-center hover:border-main transition-colors"
			>
				{currentSrc ? (
					<img src={currentSrc} alt="Avatar" className="size-full object-cover" />
				) : (
					<Upload className="size-6 text-foreground/40" />
				)}
			</div>

			<input
				ref={inputRef}
				type="file"
				accept={ALLOWED_MIME.join(",")}
				className="hidden"
				onChange={handleChange}
			/>

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

			{(clientError ?? uploadMutation.error) && (
				<Alert variant="destructive">
					<AlertDescription>
						{clientError ?? (uploadMutation.error instanceof Error ? uploadMutation.error.message : "Upload failed")}
					</AlertDescription>
				</Alert>
			)}
		</div>
	)
}
