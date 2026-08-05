import { isHeicByNameOrType } from "#/lib/image-constraints";

/**
 * Browsers cannot decode HEIC, so a picked .heic file fails in the crop dialog
 * (`new Image()` rejects) and cannot be previewed. We convert it to JPEG in the
 * browser first, using libheif compiled to WASM.
 *
 * The `/csp` entry point is required, not optional: the default build calls
 * `new Function`, which the app's `script-src 'self'` policy blocks. The CSP-safe
 * build needs only `worker-src blob:`, which spa.go grants.
 */

/** Intermediate quality — the crop step re-encodes, so limit generational loss. */
const INTERMEDIATE_QUALITY = 0.92;

/**
 * True when the file is really HEIC.
 *
 * Checks name/type first so the ~2.9 MB decoder is never downloaded for an
 * ordinary JPEG, then confirms via magic bytes.
 */
export async function isHeicFile(file: File): Promise<boolean> {
	if (!isHeicByNameOrType(file)) return false;

	try {
		const { isHeic } = await import("heic-to/csp");
		return await isHeic(file);
	} catch {
		// If the decoder can't load, treat it as HEIC so the caller surfaces a
		// conversion error rather than handing an undecodable file to the cropper.
		return true;
	}
}

/**
 * Converts a HEIC file to a JPEG File, preserving the base filename.
 *
 * Loaded on demand: a static import would add ~2.9 MB to the initial bundle.
 */
export async function convertHeicToJpeg(file: File): Promise<File> {
	let blob: Blob;

	try {
		const { heicTo } = await import("heic-to/csp");
		blob = await heicTo({
			blob: file,
			type: "image/jpeg",
			quality: INTERMEDIATE_QUALITY,
		});
	} catch {
		throw new Error("Couldn't read this HEIC image.");
	}

	const base = file.name.replace(/\.[^./\\]+$/, "") || "image";

	return new File([blob], `${base}.jpg`, { type: "image/jpeg" });
}
