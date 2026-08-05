/**
 * Single source of truth for which images the SPA accepts.
 *
 * The server allow-list (internal/files/service.go) stays at jpeg/png/gif/webp:
 * HEIC never reaches it, because the browser converts HEIC to JPEG before the
 * crop step and uploads are always re-encoded as JPEG.
 */

export const HEIC_MIME = ["image/heic", "image/heif"] as const;
export const HEIC_EXT = [".heic", ".heif"] as const;

/** Image types a user may pick, including the HEIC family handled client-side. */
export const ALLOWED_IMAGE_MIME = [
	"image/jpeg",
	"image/png",
	"image/gif",
	"image/webp",
	...HEIC_MIME,
];

/**
 * Value for an <input type="file"> accept attribute.
 *
 * Includes bare extensions as well as MIME types: Windows/Chrome commonly
 * reports an empty `file.type` for .heic, and a MIME-only accept list hides
 * those files in the picker entirely.
 */
export const FILE_INPUT_ACCEPT = [...ALLOWED_IMAGE_MIME, ...HEIC_EXT].join(",");

/** Human-readable list for validation messages. */
export const ALLOWED_IMAGE_LABEL = "JPEG, PNG, GIF, WebP, or HEIC";

function hasExtension(name: string, extensions: readonly string[]): boolean {
	const lower = name.toLowerCase();
	return extensions.some((ext) => lower.endsWith(ext));
}

/** True when the file looks like a HEIC/HEIF image by MIME type or extension. */
export function isHeicByNameOrType(file: File): boolean {
	return (
		(HEIC_MIME as readonly string[]).includes(file.type) ||
		hasExtension(file.name, HEIC_EXT)
	);
}

/**
 * True when the file is an image type we accept.
 *
 * Falls back to the extension when `file.type` is empty rather than rejecting —
 * browsers frequently report no MIME type for HEIC.
 */
export function isAcceptedImageFile(file: File): boolean {
	if (file.type) return ALLOWED_IMAGE_MIME.includes(file.type);

	return hasExtension(file.name, [
		...HEIC_EXT,
		".jpg",
		".jpeg",
		".png",
		".gif",
		".webp",
	]);
}
