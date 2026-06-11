// Journal timeline: groups entries by month/year, shows date dot, time, title, preview, people chips, label chips
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

function formatDay(dateStr: string): string {
	const d = new Date(dateStr);
	return d.toLocaleDateString(undefined, { weekday: "short", day: "numeric" });
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
					<h2 className="text-[11px] font-semibold uppercase tracking-widest text-zinc-400 mb-4 pb-2 border-b border-zinc-100">
						{group.label}
					</h2>
					<div className="relative">
						{/* vertical line */}
						<div
							className="absolute left-[7px] top-2 bottom-2 w-px bg-zinc-200"
							aria-hidden
						/>
						<ul className="space-y-5">
							{group.items.map((entry) => (
								<li key={entry.id} className="flex gap-4 pl-6 relative">
									{/* dot */}
									<span className="absolute left-0 top-[6px] size-[15px] rounded-full border-2 border-zinc-300 bg-white" />
									<div className="flex-1 min-w-0">
										<div className="flex flex-wrap items-baseline gap-x-3 gap-y-0.5 mb-1">
											<span className="font-mono text-[11px] text-zinc-400 shrink-0">
												{formatDay(entry.occurred_at_date)}
												{entry.occurred_at_time && (
													<span className="ml-1 text-zinc-300">
														· {entry.occurred_at_time}
													</span>
												)}
											</span>
										</div>
										<Link
											to="/journal/$entryId"
											params={{ entryId: String(entry.id) }}
											className="text-[14px] font-medium text-zinc-900 hover:text-indigo-600 hover:underline leading-snug"
										>
											{entry.title}
										</Link>
										{entry.content && (
											<p className="mt-0.5 text-[12px] text-zinc-500 line-clamp-2 leading-relaxed">
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
