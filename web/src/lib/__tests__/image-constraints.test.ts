import { describe, expect, it } from "vitest";
import {
	FILE_INPUT_ACCEPT,
	isAcceptedImageFile,
	isHeicByNameOrType,
} from "#/lib/image-constraints";

function file(name: string, type: string): File {
	return new File([new Uint8Array([1, 2, 3])], name, { type });
}

describe("isAcceptedImageFile", () => {
	it("accepts the standard image types", () => {
		expect(isAcceptedImageFile(file("a.jpg", "image/jpeg"))).toBe(true);
		expect(isAcceptedImageFile(file("a.png", "image/png"))).toBe(true);
		expect(isAcceptedImageFile(file("a.gif", "image/gif"))).toBe(true);
		expect(isAcceptedImageFile(file("a.webp", "image/webp"))).toBe(true);
	});

	it("accepts HEIC by MIME type", () => {
		expect(isAcceptedImageFile(file("IMG_1234.heic", "image/heic"))).toBe(true);
		expect(isAcceptedImageFile(file("IMG_1234.heif", "image/heif"))).toBe(true);
	});

	// Windows/Chrome often reports no MIME type for .heic; rejecting those would
	// make the feature fail for a large share of users.
	it("falls back to the extension when the browser reports no MIME type", () => {
		expect(isAcceptedImageFile(file("IMG_1234.heic", ""))).toBe(true);
		expect(isAcceptedImageFile(file("IMG_1234.HEIC", ""))).toBe(true);
		expect(isAcceptedImageFile(file("photo.jpeg", ""))).toBe(true);
	});

	it("rejects non-image files", () => {
		expect(isAcceptedImageFile(file("notes.txt", "text/plain"))).toBe(false);
		expect(isAcceptedImageFile(file("clip.mp4", "video/mp4"))).toBe(false);
		expect(isAcceptedImageFile(file("archive.zip", ""))).toBe(false);
	});
});

describe("isHeicByNameOrType", () => {
	it("detects HEIC by type or extension, case-insensitively", () => {
		expect(isHeicByNameOrType(file("a.heic", "image/heic"))).toBe(true);
		expect(isHeicByNameOrType(file("a.HEIF", ""))).toBe(true);
	});

	it("does not flag ordinary images", () => {
		expect(isHeicByNameOrType(file("a.jpg", "image/jpeg"))).toBe(false);
		expect(isHeicByNameOrType(file("a.png", ""))).toBe(false);
	});
});

describe("FILE_INPUT_ACCEPT", () => {
	// A MIME-only accept list hides .heic files in the OS picker on Windows.
	it("lists both HEIC MIME types and bare extensions", () => {
		expect(FILE_INPUT_ACCEPT).toContain("image/heic");
		expect(FILE_INPUT_ACCEPT).toContain(".heic");
		expect(FILE_INPUT_ACCEPT).toContain("image/jpeg");
	});
});
