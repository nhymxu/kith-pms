import { useQuery } from "@tanstack/react-query";
import { useEffect } from "react";
import { getSettings } from "#/endpoints/settings";
import { saveUserPrefs } from "#/lib/format-datetime";
import { setLanguage } from "#/lib/i18n";

// Follows settings.language (shared ["settings"] cache) so a save, or a change
// from another device, re-renders the UI live. The prefs cache gives first paint.
export function useSyncLanguage(): void {
	const { data } = useQuery({
		queryKey: ["settings"],
		queryFn: getSettings,
		staleTime: 5 * 60 * 1000,
	});
	const language = data?.language;

	useEffect(() => {
		if (!language) return;
		saveUserPrefs({ language });
		setLanguage(language).catch((err) =>
			console.warn("i18n: failed to load catalog", err),
		);
	}, [language]);
}
