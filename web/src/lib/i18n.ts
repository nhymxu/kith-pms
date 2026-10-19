import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import { z } from "zod";
import { getUserPrefs, type Language } from "#/lib/format-datetime";
import enAudit from "#/locales/en/audit.json";
import enCommon from "#/locales/en/common.json";
import enDashboard from "#/locales/en/dashboard.json";
import enDates from "#/locales/en/dates.json";
import enErrors from "#/locales/en/errors.json";
import enGifts from "#/locales/en/gifts.json";
import enJournal from "#/locales/en/journal.json";
import enMe from "#/locales/en/me.json";
import enNetwork from "#/locales/en/network.json";
import enNotes from "#/locales/en/notes.json";
import enPeople from "#/locales/en/people.json";
import enReminders from "#/locales/en/reminders.json";
import enSearch from "#/locales/en/search.json";
import enSettings from "#/locales/en/settings.json";

export const defaultNS = "common";

export const resources = {
	common: enCommon,
	people: enPeople,
	settings: enSettings,
	audit: enAudit,
	dashboard: enDashboard,
	dates: enDates,
	errors: enErrors,
	gifts: enGifts,
	journal: enJournal,
	me: enMe,
	network: enNetwork,
	notes: enNotes,
	reminders: enReminders,
	search: enSearch,
} as const;

// Non-English catalogs become lazy chunks served from /assets (same-origin,
// so the script-src 'self' CSP holds). English is bundled: it is the sync fallback.
const lazyCatalogs = import.meta.glob<Record<string, unknown>>(
	["../locales/*/*.json", "!../locales/en/*.json"],
	{ import: "default" },
);

const loaded = new Set<string>(["en"]);

async function loadCatalogs(lng: Language): Promise<void> {
	if (loaded.has(lng)) return;
	await Promise.all(
		Object.entries(lazyCatalogs)
			.filter(([path]) => path.includes(`/locales/${lng}/`))
			.map(async ([path, load]) => {
				const ns = path.slice(path.lastIndexOf("/") + 1, -".json".length);
				i18n.addResourceBundle(lng, ns, await load(), true, true);
			}),
	);
	loaded.add(lng);
}

let requested: Language = "en";

export async function setLanguage(lng: Language): Promise<void> {
	requested = lng;
	await loadCatalogs(lng);
	// A newer call may have been issued while this chunk was loading.
	if (requested !== lng) return;
	if (i18n.language !== lng) await i18n.changeLanguage(lng);
	document.documentElement.lang = lng;
	z.config(lng === "vi" ? z.locales.vi() : z.locales.en());
}

i18n.use(initReactI18next).init({
	resources: { en: resources },
	lng: "en",
	fallbackLng: "en",
	defaultNS,
	ns: Object.keys(resources),
	interpolation: { escapeValue: false },
});

// Awaited before first render so a cached non-English pref never flashes English.
export const i18nReady = setLanguage(getUserPrefs().language).catch((err) => {
	// Missing chunk (offline): stay on English rather than block the app.
	console.warn("i18n: failed to load catalog", err);
	document.documentElement.lang = "en";
});

export default i18n;
