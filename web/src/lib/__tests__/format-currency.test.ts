import { afterEach, describe, expect, it } from "vitest";
import { formatCurrency, formatNumber } from "#/lib/format-currency";

afterEach(() => {
	localStorage.clear();
});

describe("formatNumber", () => {
	it("groups thousands with comma/dot by default", () => {
		expect(formatNumber(1000000.5)).toBe("1,000,000.50");
	});

	it("switches to dot-thousands/comma-decimal when preferred", () => {
		localStorage.setItem(
			"kith_user_prefs",
			JSON.stringify({ numberFormat: "1.234,56" }),
		);
		expect(formatNumber(1000000.5)).toBe("1.000.000,50");
	});

	it("preserves the negative sign", () => {
		expect(formatNumber(-42.5)).toBe("-42.50");
	});

	it("renders — instead of garbage for non-finite input", () => {
		expect(formatNumber(Number.NaN)).toBe("—");
		expect(formatNumber(Number.POSITIVE_INFINITY)).toBe("—");
	});
});

describe("formatCurrency", () => {
	it("renders — for a null amount", () => {
		expect(formatCurrency(null, "USD")).toBe("—");
	});

	it("places $ before the amount for USD", () => {
		expect(formatCurrency(100050000, "USD")).toBe("$1,000,500.00");
	});

	it("places the symbol after the amount for VND", () => {
		expect(formatCurrency(100000050, "VND")).toBe("1,000,000.50 ₫");
	});

	it("falls back to the raw currency code for unknown currencies", () => {
		expect(formatCurrency(150, "XYZ")).toBe("XYZ 1.50");
	});

	it("defaults to USD when currency is missing", () => {
		expect(formatCurrency(150, null)).toBe("$1.50");
	});

	it("renders zero amounts instead of falling through to the null case", () => {
		expect(formatCurrency(0, "USD")).toBe("$0.00");
	});

	it("disambiguates currencies that would otherwise share a bare symbol", () => {
		expect(formatCurrency(100000, "AUD")).toBe("A$1,000.00");
		expect(formatCurrency(100000, "USD")).toBe("$1,000.00");
		expect(formatCurrency(100000, "CNY")).toBe("CN¥1,000.00");
		expect(formatCurrency(100000, "JPY")).toBe("¥1,000.00");
	});

	it("is case-insensitive on the currency code", () => {
		expect(formatCurrency(100000, "vnd")).toBe("1,000.00 ₫");
	});
});
