import { afterEach, describe, expect, it } from "vitest";
import i18n, { setLanguage } from "#/lib/i18n";

afterEach(() => setLanguage("en"));

describe("i18n", () => {
	it("swaps text and <html lang> when the language changes", async () => {
		expect(i18n.t("save")).toBe("Save");
		await setLanguage("vi");
		expect(i18n.t("save")).toBe("Lưu");
		expect(document.documentElement.lang).toBe("vi");
	});

	it("keeps the latest language when calls overlap", async () => {
		const first = setLanguage("vi");
		await setLanguage("en");
		await first;
		expect(i18n.language).toBe("en");
		expect(document.documentElement.lang).toBe("en");
	});

	it("falls back to en for a key missing in vi", async () => {
		await setLanguage("vi");
		i18n.addResource("en", "common", "onlyEn", "English only");
		// @ts-expect-error key exists only at runtime in this test
		expect(i18n.t("onlyEn")).toBe("English only");
	});
});
