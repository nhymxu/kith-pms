import { describe, expect, it } from "vitest";
import { daysRemaining } from "#/lib/format-datetime";

describe("daysRemaining", () => {
	it("counts down from the retention window", () => {
		const deletedAt = new Date(Date.now() - 5 * 86_400_000).toISOString();
		expect(daysRemaining(deletedAt, 30)).toBe(25);
	});

	it("clamps to 0 once the retention window has passed", () => {
		const deletedAt = new Date(Date.now() - 45 * 86_400_000).toISOString();
		expect(daysRemaining(deletedAt, 30)).toBe(0);
	});

	it("returns the full window for a just-deleted record", () => {
		const deletedAt = new Date().toISOString();
		expect(daysRemaining(deletedAt, 30)).toBe(30);
	});
});
