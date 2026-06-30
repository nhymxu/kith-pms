// Date/time formatting utilities that respect user preferences stored in localStorage.
// localStorage is a write-through cache; the DB (via /v1/settings) is the source of truth.

export type DateFormat = "MM/DD/YYYY" | "DD/MM/YYYY" | "YYYY-MM-DD";
export type TimeFormat = "12h" | "24h";
export type NetworkColorBy = "labels" | "type";

export interface UserPrefs {
	dateFormat: DateFormat;
	timeFormat: TimeFormat;
	timezone: string;
	networkColorBy: NetworkColorBy;
	networkShowAvatar: boolean;
	networkShowOnlyMine: boolean;
	networkShowUnconnected: boolean;
}

const STORAGE_KEY = "kith_user_prefs";

const DEFAULTS: UserPrefs = {
	dateFormat: "YYYY-MM-DD",
	timeFormat: "24h",
	timezone: "UTC",
	networkColorBy: "labels",
	networkShowAvatar: false,
	networkShowOnlyMine: false,
	networkShowUnconnected: true,
};

export function getUserPrefs(): UserPrefs {
	try {
		const raw = localStorage.getItem(STORAGE_KEY);
		if (!raw) return DEFAULTS;
		return { ...DEFAULTS, ...JSON.parse(raw) };
	} catch {
		return DEFAULTS;
	}
}

export function saveUserPrefs(prefs: Partial<UserPrefs>): void {
	const current = getUserPrefs();
	localStorage.setItem(STORAGE_KEY, JSON.stringify({ ...current, ...prefs }));
}

export function getNetworkPrefs(): Pick<
	UserPrefs,
	| "networkColorBy"
	| "networkShowAvatar"
	| "networkShowOnlyMine"
	| "networkShowUnconnected"
> {
	const p = getUserPrefs();
	return {
		networkColorBy: p.networkColorBy,
		networkShowAvatar: p.networkShowAvatar,
		networkShowOnlyMine: p.networkShowOnlyMine,
		networkShowUnconnected: p.networkShowUnconnected,
	};
}

// Fetch settings from the API and seed localStorage. Called once on authenticated app load.
export async function syncSettingsFromApi(): Promise<void> {
	try {
		const { getSettings } = await import("#/endpoints/settings");
		const s = await getSettings();
		saveUserPrefs({
			dateFormat: s.date_format as DateFormat,
			timeFormat: s.time_format as TimeFormat,
			timezone: s.timezone,
			networkColorBy: s.network_color_by,
			networkShowAvatar: s.network_show_avatar,
			networkShowOnlyMine: s.network_show_only_mine,
			networkShowUnconnected: s.network_show_unconnected,
		});
	} catch {
		// Non-fatal: fall back to whatever is already in localStorage / defaults.
	}
}

// Format a date string (YYYY-MM-DD or ISO) according to user prefs.
export function formatDate(dateStr: string | null | undefined): string {
	if (!dateStr) return "—";
	// Parse as local date to avoid timezone shift on date-only strings
	const parts = dateStr.slice(0, 10).split("-");
	if (parts.length !== 3) return dateStr;
	const [y, m, d] = parts;
	const prefs = getUserPrefs();
	switch (prefs.dateFormat) {
		case "MM/DD/YYYY":
			return `${m}/${d}/${y}`;
		case "DD/MM/YYYY":
			return `${d}/${m}/${y}`;
		default:
			return `${y}-${m}-${d}`;
	}
}

// Format a time string (HH:MM or HH:MM:SS) according to user prefs.
export function formatTime(timeStr: string | null | undefined): string {
	if (!timeStr) return "";
	const [hStr, mStr] = timeStr.split(":");
	const h = Number.parseInt(hStr, 10);
	const min = mStr ?? "00";
	const prefs = getUserPrefs();
	if (prefs.timeFormat === "12h") {
		const period = h >= 12 ? "PM" : "AM";
		const h12 = h % 12 || 12;
		return `${h12}:${min} ${period}`;
	}
	return `${String(h).padStart(2, "0")}:${min}`;
}

// Format an ISO datetime string according to user prefs (date + time).
export function formatDateTime(isoStr: string | null | undefined): string {
	if (!isoStr) return "—";
	const d = new Date(isoStr);
	if (Number.isNaN(d.getTime())) return isoStr;
	const prefs = getUserPrefs();
	// Convert to user's timezone
	const parts = new Intl.DateTimeFormat("en-CA", {
		timeZone: prefs.timezone,
		year: "numeric",
		month: "2-digit",
		day: "2-digit",
		hour: "2-digit",
		minute: "2-digit",
		hour12: false,
	}).formatToParts(d);
	const get = (type: string) => parts.find((p) => p.type === type)?.value ?? "";
	const y = get("year"),
		m = get("month"),
		day = get("day");
	const h = Number.parseInt(get("hour"), 10);
	const min = get("minute");

	let datePart: string;
	switch (prefs.dateFormat) {
		case "MM/DD/YYYY":
			datePart = `${m}/${day}/${y}`;
			break;
		case "DD/MM/YYYY":
			datePart = `${day}/${m}/${y}`;
			break;
		default:
			datePart = `${y}-${m}-${day}`;
	}

	let timePart: string;
	if (prefs.timeFormat === "12h") {
		const period = h >= 12 ? "PM" : "AM";
		const h12 = h % 12 || 12;
		timePart = `${h12}:${min} ${period}`;
	} else {
		timePart = `${String(h).padStart(2, "0")}:${min}`;
	}

	return `${datePart} ${timePart}`;
}

// Convert a UTC ISO string to "YYYY-MM-DDTHH:MM" in the user's timezone for datetime-local inputs.
export function utcToDatetimeLocal(
	isoStr: string | null | undefined,
	tz?: string,
): string {
	if (!isoStr) return "";
	const d = new Date(isoStr);
	if (Number.isNaN(d.getTime())) return "";
	const timezone = tz ?? getUserPrefs().timezone;
	const parts = new Intl.DateTimeFormat("en-CA", {
		timeZone: timezone,
		year: "numeric",
		month: "2-digit",
		day: "2-digit",
		hour: "2-digit",
		minute: "2-digit",
		hour12: false,
	}).formatToParts(d);
	const get = (type: string) => parts.find((p) => p.type === type)?.value ?? "";
	const h = get("hour") === "24" ? "00" : get("hour");
	return `${get("year")}-${get("month")}-${get("day")}T${h}:${get("minute")}`;
}

// Convert a "YYYY-MM-DDTHH:MM" value (in the given timezone) back to a UTC ISO string.
export function datetimeLocalToUtc(localStr: string, tz?: string): string {
	if (!localStr) return "";
	const timezone = tz ?? getUserPrefs().timezone;
	const [datePart, timePart = "00:00"] = localStr.split("T");
	const [year, month, day] = datePart.split("-").map(Number);
	const [hours, minutes] = timePart.split(":").map(Number);

	// Treat the input as UTC to get a reference point, then find the offset.
	const refUtc = Date.UTC(year, month - 1, day, hours, minutes);
	const fmt = new Intl.DateTimeFormat("en-CA", {
		timeZone: timezone,
		year: "numeric",
		month: "2-digit",
		day: "2-digit",
		hour: "2-digit",
		minute: "2-digit",
		hour12: false,
	});
	const p = fmt.formatToParts(new Date(refUtc));
	const g = (type: string) => p.find((x) => x.type === type)?.value ?? "0";
	const tzH = g("hour") === "24" ? 0 : Number(g("hour"));
	const tzLocalMs = Date.UTC(
		Number(g("year")),
		Number(g("month")) - 1,
		Number(g("day")),
		tzH,
		Number(g("minute")),
	);
	const offsetMs = tzLocalMs - refUtc;
	return new Date(refUtc - offsetMs).toISOString();
}
