import { useCallback, useState } from "react";

const STORAGE_PREFIX = "kith.page_size.";

function readOverride(storageKey: string): number | null {
	const raw = localStorage.getItem(storageKey);
	const n = raw ? Number(raw) : Number.NaN;
	return Number.isFinite(n) && n > 0 ? n : null;
}

// Per-list page-size override, persisted in localStorage. Takes priority over
// the server-side default_page_size setting but not an explicit ?page_size=
// URL param (deep links must stay reproducible regardless of local state).
export function usePageSizeOverride(listKey: string) {
	const storageKey = `${STORAGE_PREFIX}${listKey}`;
	const [override, setOverride] = useState<number | null>(() =>
		readOverride(storageKey),
	);

	const setPageSize = useCallback(
		(n: number) => {
			localStorage.setItem(storageKey, String(n));
			setOverride(n);
		},
		[storageKey],
	);

	const clear = useCallback(() => {
		localStorage.removeItem(storageKey);
		setOverride(null);
	}, [storageKey]);

	return { override, setPageSize, clear };
}
