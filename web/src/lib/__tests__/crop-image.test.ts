import { afterEach, describe, expect, it, vi } from "vitest";
import { blobToFile, fitWithin, loadImageWithSize } from "#/lib/crop-image";

// The avatar and gift crop dialogs both feed URL.createObjectURL(blob) sources
// through loadImageWithSize → new Image(). jsdom owns Image but does no
// decoding, so exercising the load contract is done by stubbing Image and
// firing onload/onerror exactly as the browser does. This is the load a
// missing connect-src blob: in production blocks (see
// internal/api/spa/spa_test.go). simulators are own properties because Vitest
// wraps the constructor and drops the prototype.
class MockImage {
	src = "";
	onload: (() => void) | null = null;
	onerror: (() => void) | null = null;
	naturalWidth = 0;
	naturalHeight = 0;
	/** Fires the load cycle, providing the decoded dimensions. */
	simulateLoad(width = 640, height = 400) {
		this.naturalWidth = width;
		this.naturalHeight = height;
		if (this.onload) this.onload();
	}
	/** Fires the failure cycle, as a CSP-blocked or invalid source would. */
	simulateError() {
		if (this.onerror) this.onerror();
	}
}

/** The Image instance the module constructed during the current test. */
let mockImageEl: MockImage | null = null;
function trackMockImage() {
	mockImageEl = new MockImage();
	return mockImageEl as unknown as typeof Image;
}

afterEach(() => {
	vi.restoreAllMocks();
	mockImageEl = null;
});

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

describe("blob source load", () => {
	// Both crop dialogs (avatar: aspectRatio=1, gift: freeform) are fed
	// URL.createObjectURL(prepared) as src and decoded through loadImageWithSize.
	it("resolves a load from a blob: src the way the crop dialogs use it", async () => {
		vi.spyOn(globalThis, "Image").mockImplementation(trackMockImage);

		const loadPromise = loadImageWithSize("blob:https://kith.app/abc-123");
		expect(mockImageEl).not.toBeNull();
		mockImageEl?.simulateLoad(640, 400);

		const { width, height } = await loadPromise;
		expect(width).toBe(640);
		expect(height).toBe(400);
	});

	it("rejects an unresolvable load instead of hanging on a broken source", async () => {
		vi.spyOn(globalThis, "Image").mockImplementation(trackMockImage);

		const loadPromise = loadImageWithSize(
			"https://invalid.invalid/missing.jpg",
		);
		mockImageEl?.simulateError();
		await expect(loadPromise).rejects.toThrow("Couldn't load image");
	});
});
