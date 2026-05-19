import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { Plus } from "lucide-react";
import { useState } from "react";
import { Button } from "#/components/ui/button";
import { listJournal } from "#/endpoints/journal";
import { keys } from "#/query-keys";
import { QuickJournalDialog } from "./quick-actions";

function SectionHeading({ children }: { children: React.ReactNode }) {
	return (
		<h2 className="text-[11px] font-semibold uppercase tracking-widest text-zinc-400 mb-2">
			{children}
		</h2>
	);
}

interface JournalSectionProps {
	personId: number;
}

export function JournalSection({ personId }: JournalSectionProps) {
	const [journalOpen, setJournalOpen] = useState(false);
	const { data } = useQuery({
		queryKey: keys.journal.list({ person_ids: [personId] }),
		queryFn: () => listJournal({ person_ids: [personId], page_size: 20 }),
	});
	const entries = data?.items ?? [];

	return (
		<div>
			<div className="flex items-center justify-between mb-2">
				<SectionHeading>Journal</SectionHeading>
				<Button
					variant="neutral"
					size="sm"
					onClick={() => setJournalOpen(true)}
				>
					<Plus className="size-3" /> Quick journal
				</Button>
			</div>
			{entries.length === 0 ? (
				<p className="text-sm text-zinc-400">
					No journal entries for this person.
				</p>
			) : (
				<div className="space-y-2">
					{entries.map((e) => (
						<Link
							key={e.id}
							to="/journal/$entryId"
							params={{ entryId: String(e.id) }}
							className="block p-2 border border-zinc-200 rounded-md hover:bg-zinc-50 text-sm"
						>
							<div className="flex items-center gap-2">
								<span className="font-medium flex-1">{e.title}</span>
								<span className="text-zinc-400 text-xs">
									{e.occurred_at_date}
								</span>
							</div>
							{e.people.length > 1 && (
								<div className="flex gap-1 mt-1 flex-wrap">
									{e.people
										.filter((p) => p.person_id !== personId)
										.map((p) => (
											<span
												key={p.person_id}
												className="text-[11px] text-zinc-400"
											>
												{p.name}
											</span>
										))}
								</div>
							)}
						</Link>
					))}
				</div>
			)}
			<QuickJournalDialog
				personId={personId}
				open={journalOpen}
				onClose={() => setJournalOpen(false)}
			/>
		</div>
	);
}
