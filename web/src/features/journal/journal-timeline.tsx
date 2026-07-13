// Journal timeline: groups entries by month/year; each entry has a 64px mono date gutter
// (weekday/day/time) + rail dot + content column (title, preview, label/people chips)
import { Link } from "@tanstack/react-router";
import { LabelChip, PersonChip } from "#/features/journal/person-label-chip";
import type { JournalActivity } from "#/schemas/journal";

interface JournalTimelineProps {
	data: JournalActivity[];
}

function formatMonthYear(dateStr: string): string {
	const d = new Date(dateStr);
	return d.toLocaleDateString(undefined, { month: "long", year: "numeric" });
}

function formatWeekday(dateStr: string): string {
	const d = new Date(dateStr);
	return d.toLocaleDateString(undefined, { weekday: "short" });
}

function formatDayNum(dateStr: string): string {
	const d = new Date(dateStr);
	return d.toLocaleDateString(undefined, { day: "numeric" });
}

function groupByMonth(
	entries: JournalActivity[],
): { label: string; items: JournalActivity[] }[] {
	const map = new Map<string, JournalActivity[]>();
	for (const entry of entries) {
		const key = formatMonthYear(entry.occurred_at_date);
		const group = map.get(key);
		if (group) group.push(entry);
		else map.set(key, [entry]);
	}
	return Array.from(map.entries()).map(([label, items]) => ({ label, items }));
}

export function JournalTimeline({ data }: JournalTimelineProps) {
	if (!data.length) {
		return (
			<p className="text-sm text-foreground/50 py-8 text-center">
				No journal entries yet.
			</p>
		);
	}

	const groups = groupByMonth(data);

	return (
		<div className="space-y-8">
			{groups.map((group) => (
				<div key={group.label}>
					<h2 className="text-[11px] font-semibold uppercase tracking-widest text-sub mb-4 pb-2 border-b border-line-soft">
						{group.label}
					</h2>
					<div className="relative">
						{/* vertical line, offset past the 64px date gutter + gap */}
						<div
							className="absolute left-[calc(5rem+7px)] top-2 bottom-2 w-px bg-line"
							aria-hidden
						/>
						<ul className="space-y-5">
							{group.items.map((entry) => (
								<li key={entry.id} className="flex gap-4 relative">
									{/* date gutter */}
									<div className="w-16 shrink-0 text-right font-mono leading-tight pt-0.5">
										<div className="text-[10px] uppercase tracking-wide text-sub">
											{formatWeekday(entry.occurred_at_date)}
										</div>
										<div className="text-[15px] font-semibold text-ink">
											{formatDayNum(entry.occurred_at_date)}
										</div>
										{entry.occurred_at_time && (
											<div className="text-[10px] text-sub">
												{entry.occurred_at_time}
											</div>
										)}
									</div>
									{/* dot */}
									<span className="absolute left-20 top-[6px] size-[15px] rounded-full border-2 border-line bg-panel" />
									<div className="flex-1 min-w-0 pl-6">
										<Link
											to="/journal/$entryId"
											params={{ entryId: String(entry.id) }}
											className="text-[14px] font-medium text-ink hover:text-accent-text hover:underline leading-snug"
										>
											{entry.title}
										</Link>
										{entry.content && (
											<p className="mt-0.5 text-[12px] text-sub line-clamp-2 leading-relaxed">
												{entry.content}
											</p>
										)}
										{entry.labels.length > 0 && (
											<div className="flex flex-wrap gap-1 mt-1.5">
												{entry.labels.map((l) => (
													<LabelChip key={l.id} label={l} />
												))}
											</div>
										)}
										{entry.people.length > 0 && (
											<div className="flex flex-wrap gap-1.5 mt-2">
												{entry.people.map((p) => (
													<PersonChip key={p.person_id} p={p} />
												))}
											</div>
										)}
									</div>
								</li>
							))}
						</ul>
					</div>
				</div>
			))}
		</div>
	);
}
