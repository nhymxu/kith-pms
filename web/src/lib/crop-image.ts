import type { Area } from "react-easy-crop";

function loadImage(src: string): Promise<HTMLImageElement> {
	return new Promise((resolve, reject) => {
		const img = new Image();
		img.onload = () => resolve(img);
		img.onerror = reject;
		img.src = src;
	});
}

/**
 * Draws the cropped region of `imageSrc` onto a canvas and returns it as a PNG blob.
 * Always re-encodes to PNG — an animated GIF source is flattened to its first frame.
 */
export async function cropImageToBlob(
	imageSrc: string,
	cropArea: Area,
): Promise<Blob> {
	const image = await loadImage(imageSrc);
	const canvas = document.createElement("canvas");
	canvas.width = Math.round(cropArea.width);
	canvas.height = Math.round(cropArea.height);

	const ctx = canvas.getContext("2d");
	if (!ctx) throw new Error("Canvas context unavailable");

	ctx.drawImage(
		image,
		cropArea.x,
		cropArea.y,
		cropArea.width,
		cropArea.height,
		0,
		0,
		cropArea.width,
		cropArea.height,
	);

	return new Promise((resolve, reject) => {
		canvas.toBlob((blob) => {
			if (blob) resolve(blob);
			else reject(new Error("Failed to encode cropped image"));
		}, "image/png");
	});
}

/** Wraps a cropped blob as a File, reusing the original name with a .png extension. */
export function blobToFile(blob: Blob, originalName: string): File {
	const base = originalName.replace(/\.[^./\\]+$/, "");
	return new File([blob], `${base}.png`, { type: "image/png" });
}
