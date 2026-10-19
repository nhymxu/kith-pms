// Date display helpers specific to the relationship graph profile card / selected panel.

import i18n from "#/lib/i18n";

export function calcAge(dob: string): number | null {
	if (dob.startsWith("--")) return null; // yearless — no year to calculate from
	const birth = new Date(dob);
	const today = new Date();
	let age = today.getFullYear() - birth.getFullYear();
	const m = today.getMonth() - birth.getMonth();
	if (m < 0 || (m === 0 && today.getDate() < birth.getDate())) age--;
	return age;
}

export function formatBirthdayLabel(dob: string): string {
	const yearless = dob.startsWith("--");
	// Reconstruct a parseable date (use leap year for --MM-DD to handle Feb 29)
	const parseable = yearless ? `2024${dob.slice(1)}` : dob;
	const d = new Date(parseable);
	if (Number.isNaN(d.getTime())) return dob;
	const monthDay = d.toLocaleDateString("en-US", {
		month: "long",
		day: "numeric",
	});
	if (yearless) return monthDay;
	const age = calcAge(dob);
	return age !== null
		? i18n.t("network:date.birthdayWithAge", {
				monthDay,
				year: d.getFullYear(),
				age,
			})
		: monthDay;
}

export function formatRelativeDate(isoStr: string): string {
	const then = new Date(isoStr);
	if (Number.isNaN(then.getTime())) return isoStr;
	const days = Math.floor((Date.now() - then.getTime()) / 86_400_000);
	if (days === 0) return i18n.t("network:date.today");
	if (days === 1) return i18n.t("network:date.yesterday");
	if (days < 7) return i18n.t("network:date.daysAgo", { count: days });
	const weeks = Math.floor(days / 7);
	if (days < 30) return i18n.t("network:date.weeksAgo", { count: weeks });
	const months = Math.floor(days / 30);
	if (days < 365) return i18n.t("network:date.monthsAgo", { count: months });
	const years = Math.floor(days / 365);
	return i18n.t("network:date.yearsAgo", { count: years });
}
