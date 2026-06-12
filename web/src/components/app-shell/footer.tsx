import { useQuery } from "@tanstack/react-query";
import { getAppInfo } from "#/endpoints/app";
import { getSettings } from "#/endpoints/settings";
import { keys } from "#/query-keys";

function useNavLoadMs(): number | null {
	if (typeof performance === "undefined") return null;
	const entries = performance.getEntriesByType("navigation");
	if (!entries.length) return null;
	const nav = entries[0] as PerformanceNavigationTiming;
	return Math.round(nav.duration);
}

export function Footer() {
	const { data: appInfo } = useQuery({
		queryKey: keys.app.info(),
		queryFn: getAppInfo,
		staleTime: Number.POSITIVE_INFINITY,
	});

	const { data: settings } = useQuery({
		queryKey: ["settings"],
		queryFn: getSettings,
		staleTime: 5 * 60 * 1000,
	});

	const loadMs = useNavLoadMs();

	return (
		<footer className="border-t border-border bg-secondary px-4 sm:px-6 py-2 flex items-center justify-between text-[11px] text-muted-foreground font-mono">
			<div className="flex items-center gap-4">
				{appInfo && (
					<span>
						Version:{" "}
						<span className="text-main font-semibold">{appInfo.version}</span>
					</span>
				)}
				{loadMs !== null && <span>Load: {loadMs}ms</span>}
			</div>
			<div>{settings && <span>Timezone: {settings.timezone}</span>}</div>
		</footer>
	);
}
