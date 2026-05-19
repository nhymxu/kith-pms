import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { BookOpen } from "lucide-react";
import { listJournal } from "#/endpoints/journal";
import { keys } from "#/query-keys";
import type { JournalActivity } from "#/schemas/journal";

function JournalCard({ entry }: { entry: JournalActivity }) {
	return (
		<Link
			to="/journal/$entryId"
			params={{ entryId: String(entry.id) }}
			className="block py-3 border-b border-zinc-100 last:border-b-0 hover:bg-zinc-50 -mx-4 px-4 transition-colors"
		>
			<div className="flex items-start justify-between gap-2">
				<p className="text-[13px] text-zinc-900 truncate flex-1">
					{entry.title}
				</p>
				<span className="font-mono text-[10px] text-zinc-500 shrink-0">
					{entry.occurred_at_date}
				</span>
			</div>
			{entry.people.length > 0 && (
				<div className="mt-1 flex flex-wrap gap-1.5">
					{entry.people.slice(0, 3).map((p) => (
						<span key={p.person_id} className="text-[10px] text-indigo-600">
							@{p.name}
						</span>
					))}
					{entry.people.length > 3 && (
						<span className="text-[10px] text-zinc-400">
							+{entry.people.length - 3}
						</span>
					)}
				</div>
			)}
		</Link>
	);
}

export function RecentJournalActivity() {
	const { data, isLoading } = useQuery({
		queryKey: keys.journal.list({ page_size: 5 }),
		queryFn: () => listJournal({ page_size: 5 }),
	});

	return (
		<div className="border border-zinc-200 rounded-md bg-white">
			<div className="flex items-center gap-2 px-4 py-3 border-b border-zinc-200">
				<BookOpen className="size-3.5 text-zinc-400" />
				<p className="text-[13px] font-medium text-zinc-900">
					Recent journal entries
				</p>
			</div>
			<div className="px-4">
				{isLoading &&
					Array.from({ length: 3 }).map((_, i) => (
						<div
							key={i}
							className="h-12 bg-zinc-100 animate-pulse rounded my-2"
						/>
					))}
				{!isLoading && !data?.items.length && (
					<p className="text-[13px] text-zinc-500 py-6 text-center">
						No journal entries yet.
					</p>
				)}
				{data?.items.map((entry) => (
					<JournalCard key={entry.id} entry={entry} />
				))}
			</div>
		</div>
	);
}
