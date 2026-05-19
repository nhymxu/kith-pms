// Date/time formatting utilities that respect user preferences stored in localStorage.

export type DateFormat = "MM/DD/YYYY" | "DD/MM/YYYY" | "YYYY-MM-DD"
export type TimeFormat = "12h" | "24h"

export interface UserPrefs {
	dateFormat: DateFormat
	timeFormat: TimeFormat
	timezone: string
}

const STORAGE_KEY = "kith_user_prefs"

const DEFAULTS: UserPrefs = {
	dateFormat: "YYYY-MM-DD",
	timeFormat: "24h",
	timezone: "UTC",
}

export function getUserPrefs(): UserPrefs {
	try {
		const raw = localStorage.getItem(STORAGE_KEY)
		if (!raw) return DEFAULTS
		return { ...DEFAULTS, ...JSON.parse(raw) }
	} catch {
		return DEFAULTS
	}
}

export function saveUserPrefs(prefs: Partial<UserPrefs>): void {
	const current = getUserPrefs()
	localStorage.setItem(STORAGE_KEY, JSON.stringify({ ...current, ...prefs }))
}

// Format a date string (YYYY-MM-DD or ISO) according to user prefs.
export function formatDate(dateStr: string | null | undefined): string {
	if (!dateStr) return "—"
	// Parse as local date to avoid timezone shift on date-only strings
	const parts = dateStr.slice(0, 10).split("-")
	if (parts.length !== 3) return dateStr
	const [y, m, d] = parts
	const prefs = getUserPrefs()
	switch (prefs.dateFormat) {
		case "MM/DD/YYYY": return `${m}/${d}/${y}`
		case "DD/MM/YYYY": return `${d}/${m}/${y}`
		case "YYYY-MM-DD":
		default: return `${y}-${m}-${d}`
	}
}

// Format a time string (HH:MM or HH:MM:SS) according to user prefs.
export function formatTime(timeStr: string | null | undefined): string {
	if (!timeStr) return ""
	const [hStr, mStr] = timeStr.split(":")
	const h = Number.parseInt(hStr, 10)
	const min = mStr ?? "00"
	const prefs = getUserPrefs()
	if (prefs.timeFormat === "12h") {
		const period = h >= 12 ? "PM" : "AM"
		const h12 = h % 12 || 12
		return `${h12}:${min} ${period}`
	}
	return `${String(h).padStart(2, "0")}:${min}`
}
