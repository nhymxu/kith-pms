import { describe, expect, it } from "vitest";
import { blobToFile, fitWithin } from "#/lib/crop-image";

describe("fitWithin", () => {
	it("scales the longest edge down to the cap, preserving aspect ratio", () => {
		expect(fitWithin(3000, 2000, 1600)).toEqual({ width: 1600, height: 1067 });
		expect(fitWithin(2000, 3000, 1600)).toEqual({ width: 1067, height: 1600 });
	});

	it("never exceeds the cap on either edge", () => {
		for (const [w, h] of [
			[4032, 3024],
			[3024, 4032],
			[5000, 5000],
			[1601, 900],
		]) {
			const out = fitWithin(w, h, 1600);
			expect(Math.max(out.width, out.height)).toBeLessThanOrEqual(1600);
		}
	});

	it("never upscales a crop smaller than the cap", () => {
		expect(fitWithin(800, 600, 1600)).toEqual({ width: 800, height: 600 });
		expect(fitWithin(1, 1, 1600)).toEqual({ width: 1, height: 1 });
	});

	it("keeps a square square", () => {
		expect(fitWithin(2400, 2400, 1024)).toEqual({ width: 1024, height: 1024 });
	});

	// Guards against a zero-sized canvas, which makes canvas.toBlob return null.
	it("never returns a zero dimension for extreme ratios", () => {
		const out = fitWithin(4000, 3, 100);
		expect(out.width).toBeGreaterThanOrEqual(1);
		expect(out.height).toBeGreaterThanOrEqual(1);
	});
});

describe("blobToFile", () => {
	const blob = new Blob([new Uint8Array([1, 2, 3])], { type: "image/jpeg" });

	it("swaps the extension to .jpg and sets the JPEG type", () => {
		const f = blobToFile(blob, "IMG_1234.heic");
		expect(f.name).toBe("IMG_1234.jpg");
		expect(f.type).toBe("image/jpeg");
	});

	it("falls back to a placeholder name when the source has none", () => {
		expect(blobToFile(blob, ".png").name).toBe("image.jpg");
		expect(blobToFile(blob, "").name).toBe("image.jpg");
	});
});
