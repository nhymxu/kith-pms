export const GIFT_IMAGE_ALLOWED_MIME = [
	"image/jpeg",
	"image/png",
	"image/gif",
	"image/webp",
];

/** react-easy-crop requires a fixed viewport ratio; 4:3 fits typical gift photos well. */
export const GIFT_IMAGE_ASPECT = 4 / 3;

export const GIFT_IMAGE_MAX_BYTES = 5 * 1024 * 1024; // 5 MB, mirrors backend limit
