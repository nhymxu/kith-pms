import { useSuspenseQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { ExternalLink, Plus } from "lucide-react";
import { useState } from "react";
import { QueryBoundary } from "#/components/query-boundary";
import { Button } from "#/components/ui/button";
import { listJournal } from "#/endpoints/journal";
import { PersonChip } from "#/features/journal/person-label-chip";
import { keys } from "#/query-keys";
import { QuickJournalDialog } from "./quick-actions";

function SectionHeading({ children }: { children: React.ReactNode }) {
	return (
		<h2 className="text-[11px] font-semibold uppercase tracking-widest text-sub mb-2">
			{children}
		</h2>
	);
}

interface JournalListInnerProps {
	personId: number;
}

function JournalListInner({ personId }: JournalListInnerProps) {
	const { data } = useSuspenseQuery({
		queryKey: keys.journal.list({ person_ids: [personId] }),
		queryFn: () => listJournal({ person_ids: [personId], page_size: 20 }),
	});
	const entries = data.items;

	if (entries.length === 0) {
		return (
			<p className="text-sm text-sub">No journal entries for this person.</p>
		);
	}

	return (
		<div className="space-y-2">
			{entries.map((e) => (
				<Link
					key={e.id}
					to="/journal/$entryId"
					params={{ entryId: String(e.id) }}
					className="block p-2 border-bw border-line rounded-md hover:bg-chip text-sm"
				>
					<div className="flex items-center gap-2">
						<span className="font-medium flex-1">{e.title}</span>
						<span className="text-sub text-xs">{e.occurred_at_date}</span>
					</div>
					{e.content && (
						<p className="text-sub text-xs mt-1 line-clamp-2">
							{e.content.length > 100
								? `${e.content.slice(0, 100)}…`
								: e.content}
						</p>
					)}
					{e.people.length > 1 && (
						<div className="flex gap-1 mt-1 flex-wrap">
							{e.people
								.filter((p) => p.person_id !== personId)
								.map((p) => (
									<PersonChip key={p.person_id} p={p} />
								))}
						</div>
					)}
				</Link>
			))}
		</div>
	);
}

interface JournalSectionProps {
	personId: number;
}

export function JournalSection({ personId }: JournalSectionProps) {
	const [journalOpen, setJournalOpen] = useState(false);

	return (
		<div>
			<div className="flex items-center justify-between mb-2">
				<SectionHeading>Journal</SectionHeading>
				<div className="flex items-center gap-5">
					<Link
						to="/journal"
						search={{ people: [personId] }}
						className="flex items-center gap-1 text-xs font-medium text-sub hover:text-ink"
					>
						View all <ExternalLink className="size-3" />
					</Link>
					<Button
						variant="neutral"
						size="sm"
						onClick={() => setJournalOpen(true)}
					>
						<Plus className="size-3" /> Quick journal
					</Button>
				</div>
			</div>
			<QueryBoundary>
				<JournalListInner personId={personId} />
			</QueryBoundary>
			<QuickJournalDialog
				personId={personId}
				open={journalOpen}
				onClose={() => setJournalOpen(false)}
			/>
		</div>
	);
}
