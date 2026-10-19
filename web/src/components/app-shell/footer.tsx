import { useSuspenseQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import { QueryBoundary } from "#/components/query-boundary";
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

const footerFallback = (
	<footer className="border-t border-border bg-secondary px-4 sm:px-6 py-2 h-[29px]" />
);

function FooterContent() {
	const { data: appInfo } = useSuspenseQuery({
		queryKey: keys.app.info(),
		queryFn: getAppInfo,
		staleTime: Number.POSITIVE_INFINITY,
	});

	const { data: settings } = useSuspenseQuery({
		queryKey: ["settings"],
		queryFn: getSettings,
		staleTime: 5 * 60 * 1000,
	});

	const loadMs = useNavLoadMs();
	const { t } = useTranslation();

	return (
		<footer className="border-t border-border bg-secondary px-4 sm:px-6 py-2 flex items-center justify-between text-[11px] text-muted-foreground font-mono">
			<div className="flex items-center gap-4">
				<span>
					{t("footer.version")}{" "}
					<span className="text-main font-semibold">{appInfo.version}</span>
				</span>
				{loadMs !== null && <span>{t("footer.load", { ms: loadMs })}</span>}
			</div>
			<div>
				<span>{t("footer.timezone", { timezone: settings.timezone })}</span>
			</div>
		</footer>
	);
}

export function Footer() {
	return (
		<QueryBoundary fallback={footerFallback}>
			<FooterContent />
		</QueryBoundary>
	);
}
