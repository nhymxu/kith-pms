import { afterEach, describe, expect, it } from "vitest";
import { ApiError } from "#/lib/api-client";
import { apiErrorMessage } from "#/lib/api-error-message";
import { setLanguage } from "#/lib/i18n";

afterEach(() => setLanguage("en"));

describe("apiErrorMessage", () => {
	it("maps a server code in the active language", async () => {
		const err = new ApiError(422, "current_password_incorrect", "english");
		expect(apiErrorMessage(err)).toBe("Current password is incorrect.");
		await setLanguage("vi");
		expect(apiErrorMessage(err)).toBe("Mật khẩu hiện tại không đúng.");
	});

	it("interpolates params", async () => {
		await setLanguage("vi");
		const err = new ApiError(422, "name_too_long", "x", { max: 80 });
		expect(apiErrorMessage(err)).toBe("Tên tối đa 80 ký tự.");
	});

	it("keeps the server text for an unmapped code", () => {
		const err = new ApiError(400, "brand_new_code", "Server says no");
		expect(apiErrorMessage(err)).toBe("Server says no");
	});

	it("falls back to the status text when the response had no code", async () => {
		await setLanguage("vi");
		expect(apiErrorMessage(new ApiError(502, "", "Bad Gateway"))).toBe(
			"Máy chủ không phản hồi. Vui lòng thử lại.",
		);
	});

	it("reports a dropped connection and plain errors", () => {
		expect(apiErrorMessage(new TypeError("Failed to fetch"))).toContain(
			"Can't reach the server",
		);
		expect(apiErrorMessage(new Error("boom"))).toBe("boom");
		expect(apiErrorMessage(null, "fallback")).toBe("fallback");
	});
});
