// Upcoming dates list: grouped by month, shows person name, date type, exact date + days-until badge
import { useSuspenseQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { QueryBoundary } from "#/components/query-boundary";
import { listUpcomingDates } from "#/endpoints/important-dates";
import { getAvatarUrl } from "#/endpoints/people";
import { keys } from "#/query-keys";

function daysUntil(dateStr: string): number {
	const today = new Date();
	today.setHours(0, 0, 0, 0);
	const target = new Date(dateStr);
	target.setHours(0, 0, 0, 0);
	return Math.round(
		(target.getTime() - today.getTime()) / (1000 * 60 * 60 * 24),
	);
}

function monthLabel(dateStr: string): string {
	const d = new Date(dateStr);
	return d.toLocaleDateString(undefined, { month: "long", year: "numeric" });
}

function exactDateLabel(dateStr: string): string {
	const d = new Date(dateStr);
	const currentYear = new Date().getFullYear();
	if (d.getFullYear() === currentYear) {
		return d.toLocaleDateString(undefined, { month: "short", day: "numeric" });
	}
	return d.toLocaleDateString(undefined, {
		month: "short",
		day: "numeric",
		year: "numeric",
	});
}

function DaysUntilBadge({ days, dateStr }: { days: number; dateStr: string }) {
	const { t } = useTranslation("dates");
	const exact = exactDateLabel(dateStr);
	if (days === 0)
		return (
			<div className="flex flex-col items-end gap-0.5">
				<span className="font-mono text-[11px] text-accent-text font-medium">
					{t("badge.today")}
				</span>
				<span className="text-[10px] text-sub">{exact}</span>
			</div>
		);
	if (days <= 7)
		return (
			<div className="flex flex-col items-end gap-0.5">
				<span className="font-mono text-[11px] text-warning-fg">
					{t("badge.inDays", { days })}
				</span>
				<span className="text-[10px] text-sub">{exact}</span>
			</div>
		);
	if (days <= 30)
		return (
			<div className="flex flex-col items-end gap-0.5">
				<span className="font-mono text-[11px] text-sub">
					{t("badge.inDays", { days })}
				</span>
				<span className="text-[10px] text-sub">{exact}</span>
			</div>
		);
	return (
		<div className="flex flex-col items-end gap-0.5">
			<span className="font-mono text-[11px] text-sub">
				{t("badge.inDays", { days })}
			</span>
			<span className="text-[10px] text-sub">{exact}</span>
		</div>
	);
}

function DatesListInner() {
	const { t } = useTranslation("dates");
	const { data } = useSuspenseQuery({
		queryKey: keys.dates.upcoming(),
		queryFn: () => listUpcomingDates(90),
	});

	if (!data.length)
		return <p className="text-[13px] text-sub">{t("list.empty")}</p>;

	// Group by month of next_occurrence
	const groups = new Map<string, typeof data>();
	for (const item of data) {
		const key = monthLabel(item.next_occurrence);
		if (!groups.has(key)) groups.set(key, []);
		groups.get(key)?.push(item);
	}

	return (
		<div className="space-y-6">
			{Array.from(groups.entries()).map(([month, items]) => (
				<div key={month}>
					<h2 className="text-[11px] font-medium uppercase tracking-wider text-sub mb-2">
						{month}
					</h2>
					<div className="border-bw border-line rounded-base bg-panel divide-y divide-line-soft">
						{items.map((item, _i) => {
							const days = daysUntil(item.next_occurrence);
							return (
								<div
									key={`${item.person.id}-${item.kind}`}
									className="flex items-center gap-3 px-4 py-3 hover:bg-chip transition-colors"
								>
									<div className="size-10 rounded-full overflow-hidden shrink-0 bg-chip flex items-center justify-center text-sm font-medium text-chip-fg font-mono">
										{item.person.avatar_path ? (
											<img
												src={getAvatarUrl(item.person.id)}
												alt={item.person.name}
												className="size-full object-cover"
											/>
										) : (
											<span>{item.person.name.charAt(0).toUpperCase()}</span>
										)}
									</div>
									<div className="flex-1 min-w-0">
										<Link
											to="/people/$personId"
											params={{ personId: String(item.person.id) }}
											className="text-[13px] text-ink hover:text-accent-text hover:underline"
										>
											{item.person.name}
										</Link>
										<p className="text-[11px] text-sub capitalize">
											{t(`people:sections.dates.kind.${item.kind}`, {
												defaultValue: item.kind,
											})}
											{item.years_since > 0
												? ` · ${t("list.years", { count: item.years_since + 1 })}`
												: ""}
										</p>
									</div>
									<DaysUntilBadge days={days} dateStr={item.next_occurrence} />
								</div>
							);
						})}
					</div>
				</div>
			))}
		</div>
	);
}

function DatesListFallback() {
	const { t } = useTranslation("dates");
	return <p className="text-[13px] text-sub">{t("list.loading")}</p>;
}

export function DatesList() {
	return (
		<QueryBoundary fallback={<DatesListFallback />}>
			<DatesListInner />
		</QueryBoundary>
	);
}
