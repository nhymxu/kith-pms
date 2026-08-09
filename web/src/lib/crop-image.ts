export function loadImage(src: string): Promise<HTMLImageElement> {
	return new Promise((resolve, reject) => {
		const img = new Image();
		img.onload = () => resolve(img);
		img.onerror = () => reject(new Error("Couldn't load image"));
		img.src = src;
	});
}

/** A plain source-pixel rectangle, decoupled from any cropper library. */
export type CropRect = { x: number; y: number; width: number; height: number };

/** Loads the source image, returning its decoded size alongside the element. */
export async function loadImageWithSize(src: string): Promise<{
	image: HTMLImageElement;
	width: number;
	height: number;
}> {
	const image = await loadImage(src);
	return { image, width: image.naturalWidth, height: image.naturalHeight };
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
 * Decodes `imageSrc` and returns a JPEG blob of its `rect` region, scaled so its
 * longest edge is at most `maxEdgePx`.
 *
 * `rect` is plain source-pixel geometry {x, y, width, height}, decoupled from
 * any cropper library so both the advanced-cropper dialog and the uncropped
 * (skip-crop) path can share it.
 */
export async function cropImageToBlob(
	imageSrc: string,
	rect: CropRect,
	{ maxEdgePx, quality }: CropEncodeOptions,
): Promise<Blob> {
	return encodeCropped(await loadImage(imageSrc), rect, { maxEdgePx, quality });
}

/**
 * Draws the `rect` region of an already-decoded image onto a canvas and returns
 * it as a JPEG blob, scaled so its longest edge is at most `maxEdgePx`.
 *
 * Unlike cropImageToBlob, this takes the decoded image so callers that already
 * loaded it (e.g. to read its natural size) skip a second decode.
 */
export function encodeCropped(
	image: HTMLImageElement,
	rect: CropRect,
	{ maxEdgePx, quality }: CropEncodeOptions,
): Promise<Blob> {
	// Clamp the source rect to the decoded bounds — advanced-cropper's
	// coordinates are floats and can round a whisker past the edge (e.g. the
	// hand-typed full-image rect), which drawImage would draw off-canvas/flip.
	const srcX = Math.max(0, Math.round(rect.x));
	const srcY = Math.max(0, Math.round(rect.y));
	const srcWidth = Math.min(rect.width, image.naturalWidth - srcX);
	const srcHeight = Math.min(rect.height, image.naturalHeight - srcY);

	const size = fitWithin(srcWidth, srcHeight, maxEdgePx);

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
		srcX,
		srcY,
		srcWidth,
		srcHeight,
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
