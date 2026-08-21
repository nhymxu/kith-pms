// Currency amount formatting that respects the user's number_format preference
// (grouping/decimal separators) and each currency's conventional symbol placement.

import { getUserPrefs } from "#/lib/format-datetime";

interface CurrencyMeta {
	symbol: string;
	position: "before" | "after";
}

// Symbol placement is a property of the currency itself, not a user preference —
// e.g. VND conventionally places đ/₫ after the amount while USD/EUR place their
// symbol before it. Unknown codes fall back to "CODE amount" (previous behavior).
// Symbols are disambiguated (A$, C$, S$, CN¥) rather than reused across currencies —
// a bare "$" or "¥" would make an AUD gift and a USD gift render identically.
const CURRENCY_META: Record<string, CurrencyMeta> = {
	USD: { symbol: "$", position: "before" },
	EUR: { symbol: "€", position: "before" },
	GBP: { symbol: "£", position: "before" },
	JPY: { symbol: "¥", position: "before" },
	CNY: { symbol: "CN¥", position: "before" },
	AUD: { symbol: "A$", position: "before" },
	CAD: { symbol: "C$", position: "before" },
	SGD: { symbol: "S$", position: "before" },
	INR: { symbol: "₹", position: "before" },
	KRW: { symbol: "₩", position: "before" },
	THB: { symbol: "฿", position: "before" },
	VND: { symbol: "₫", position: "after" },
};

// Format a plain number with the grouping/decimal separators from user prefs,
// always at 2 decimal places (amounts are stored as integer cents).
//
// The number_format enum's valid separator styles are duplicated across:
// internal/settings/service.go (validNumberFormats), web/src/schemas/settings.ts,
// the NumberFormat union + DEFAULTS in web/src/lib/format-datetime.ts,
// NUMBER_FORMAT_OPTIONS in web/src/routes/_authed/settings/_layout.general.tsx,
// and the branch below. Adding a style must update all five or it will pass
// validation and then silently render with the default separators.
export function formatNumber(value: number): string {
	if (!Number.isFinite(value)) return "—";

	const { numberFormat } = getUserPrefs();
	const sign = value < 0 ? "-" : "";
	const [intPart, decPart] = Math.abs(value).toFixed(2).split(".");
	const grouped = intPart.replace(/\B(?=(\d{3})+(?!\d))/g, ",");

	if (numberFormat === "1.234,56") {
		return `${sign}${grouped.replace(/,/g, ".")},${decPart}`;
	}
	return `${sign}${grouped}.${decPart}`;
}

// Format an amount stored as integer cents alongside its currency code.
export function formatCurrency(
	amountCents: number | null | undefined,
	currency: string | null | undefined,
): string {
	if (amountCents == null) return "—";

	const code = (currency || "USD").toUpperCase();
	const amount = formatNumber(amountCents / 100);
	const meta = CURRENCY_META[code];

	if (!meta) return `${code} ${amount}`;
	return meta.position === "before"
		? `${meta.symbol}${amount}`
		: `${amount} ${meta.symbol}`;
}
