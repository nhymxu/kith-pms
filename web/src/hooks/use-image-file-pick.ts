import { useRef, useState } from "react";
import { convertHeicToJpeg, isHeicFile } from "#/lib/heic-convert";
import {
	ALLOWED_IMAGE_LABEL,
	isAcceptedImageFile,
} from "#/lib/image-constraints";

interface UseImageFilePickOptions {
	/** Upload cap in bytes, from server config. */
	maxBytes: number;
	/** Same cap in MB, used only for the error message. */
	maxSizeMB: number;
}

/**
 * Shared pick-and-prepare step for the image upload surfaces.
 *
 * Validates the chosen file, then converts HEIC to JPEG so the crop dialog can
 * decode it — browsers cannot render HEIC, and the cropper would otherwise fail
 * with a misleading "couldn't crop this image" error.
 *
 * Returns the file to hand to the cropper, or null when it was rejected (in
 * which case `error` explains why).
 */
export function useImageFilePick({
	maxBytes,
	maxSizeMB,
}: UseImageFilePickOptions) {
	const [isConverting, setIsConverting] = useState(false);
	const [error, setError] = useState<string | null>(null);

	// Synchronous latch. State alone is not enough: the HEIC sniff dynamically
	// imports a ~2.9 MB chunk, so several seconds can pass before isConverting
	// re-renders the inputs as disabled — long enough for a second pick (or a
	// drop, which ignores `disabled`) to start a concurrent conversion and leak
	// an object URL.
	const inFlight = useRef(false);

	async function prepare(file: File): Promise<File | null> {
		if (inFlight.current) return null;

		inFlight.current = true;
		setError(null);

		try {
			if (!isAcceptedImageFile(file)) {
				setError(`Only ${ALLOWED_IMAGE_LABEL} images are allowed.`);
				return null;
			}

			if (file.size > maxBytes) {
				setError(`File must be under ${maxSizeMB} MB.`);
				return null;
			}

			if (!(await isHeicFile(file))) return file;

			setIsConverting(true);
			try {
				return await convertHeicToJpeg(file);
			} catch (err) {
				setError(
					err instanceof Error ? err.message : "Couldn't read this image.",
				);
				return null;
			} finally {
				setIsConverting(false);
			}
		} finally {
			inFlight.current = false;
		}
	}

	return { prepare, isConverting, error, setError };
}
