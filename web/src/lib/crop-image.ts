import type { Area } from "react-easy-crop";

function loadImage(src: string): Promise<HTMLImageElement> {
	return new Promise((resolve, reject) => {
		const img = new Image();
		img.onload = () => resolve(img);
		img.onerror = reject;
		img.src = src;
	});
}

export type CropEncodeOptions = {
	/** Longest edge of the output; larger crops are scaled down, never up. */
	maxEdgePx: number;
	/** JPEG quality as an integer 1-100 (server config units). */
	quality: number;
};

/**
 * Scales a crop region so its longest edge is at most `maxEdgePx`, preserving
 * aspect ratio. Never upscales — a small crop stays at its native size.
 *
 * Extracted from cropImageToBlob so it is testable: the canvas path cannot run
 * under jsdom.
 */
export function fitWithin(
	width: number,
	height: number,
	maxEdgePx: number,
): { width: number; height: number } {
	const srcWidth = Math.max(1, Math.round(width));
	const srcHeight = Math.max(1, Math.round(height));
	const scale = Math.min(1, maxEdgePx / Math.max(srcWidth, srcHeight));

	return {
		width: Math.max(1, Math.round(srcWidth * scale)),
		height: Math.max(1, Math.round(srcHeight * scale)),
	};
}

/**
 * Draws the cropped region of `imageSrc` onto a canvas and returns it as a JPEG
 * blob, scaled so its longest edge is at most `maxEdgePx`.
 *
 * JPEG rather than PNG: for photographic content PNG runs ~8.6x larger, which
 * pushed phone-sized crops past the upload cap. An animated GIF source is
 * flattened to its first frame, and transparency is lost — hence the white fill
 * below, since a bare canvas is transparent-black and JPEG would flatten that
 * to black.
 */
export async function cropImageToBlob(
	imageSrc: string,
	cropArea: Area,
	{ maxEdgePx, quality }: CropEncodeOptions,
): Promise<Blob> {
	const image = await loadImage(imageSrc);

	const size = fitWithin(cropArea.width, cropArea.height, maxEdgePx);

	const canvas = document.createElement("canvas");
	canvas.width = size.width;
	canvas.height = size.height;

	const ctx = canvas.getContext("2d");
	if (!ctx) throw new Error("Canvas context unavailable");

	ctx.imageSmoothingEnabled = true;
	ctx.imageSmoothingQuality = "high";
	ctx.fillStyle = "#ffffff";
	ctx.fillRect(0, 0, canvas.width, canvas.height);

	ctx.drawImage(
		image,
		cropArea.x,
		cropArea.y,
		cropArea.width,
		cropArea.height,
		0,
		0,
		canvas.width,
		canvas.height,
	);

	return new Promise((resolve, reject) => {
		canvas.toBlob(
			(blob) => {
				if (blob) resolve(blob);
				else reject(new Error("Failed to encode cropped image"));
			},
			"image/jpeg",
			// canvas.toBlob wants 0-1; it silently ignores out-of-range values.
			Math.min(100, Math.max(1, quality)) / 100,
		);
	});
}

/** Wraps a cropped blob as a File, reusing the original name with a .jpg extension. */
export function blobToFile(blob: Blob, originalName: string): File {
	const base = originalName.replace(/\.[^./\\]+$/, "") || "image";
	return new File([blob], `${base}.jpg`, { type: "image/jpeg" });
}
