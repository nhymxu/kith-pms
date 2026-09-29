import {
	keepPreviousData,
	useQuery,
	useSuspenseQuery,
} from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { useState } from "react";
import { QueryBoundary } from "#/components/query-boundary";
import { listAllNotes } from "#/endpoints/notes";
import { getAvatarUrl, listPeople } from "#/endpoints/people";
import { JournalPagination } from "#/features/journal/journal-pagination";
import { NotesDateRange } from "#/features/notes/notes-date-range";
import { formatDateTime } from "#/lib/format-datetime";
import { keys } from "#/query-keys";
import type { NoteWithPerson } from "#/schemas/note";

function groupByMonth(notes: NoteWithPerson[]) {
	const map = new Map<string, NoteWithPerson[]>();
	for (const n of notes) {
		const label = new Date(n.created_at).toLocaleDateString(undefined, {
			month: "long",
			year: "numeric",
		});
		const group = map.get(label);
		if (group) group.push(n);
		else map.set(label, [n]);
	}
	return Array.from(map, ([label, items]) => ({ label, items }));
}

const PAGE_SIZE = 20;

function AllNotesListInner() {
	const [page, setPage] = useState(1);
	const [fromDate, setFromDate] = useState("");
	const [toDate, setToDate] = useState("");
	const [personIds, setPersonIds] = useState<number[]>([]);

	const filters = {
		person_ids: personIds.length ? personIds : undefined,
		from_date: fromDate || undefined,
		to_date: toDate || undefined,
	};

	const { data: people } = useSuspenseQuery({
		queryKey: keys.people.list({ page_size: 500 }),
		queryFn: () => listPeople({ page_size: 500 }),
	});

	const { data } = useQuery({
		queryKey: keys.notes.list({
			scope: "all",
			page,
			page_size: PAGE_SIZE,
			...filters,
		}),
		queryFn: () => listAllNotes({ page, page_size: PAGE_SIZE, ...filters }),
		placeholderData: keepPreviousData,
	});

	function change<T>(set: (v: T) => void) {
		return (v: T) => {
			set(v);
			setPage(1);
		};
	}

	const filterUi = (
		<div className="space-y-3 mb-6">
			<NotesDateRange
				idPrefix="notes"
				from={fromDate}
				to={toDate}
				onChange={(f, t) => {
					setFromDate(f);
					setToDate(t);
					setPage(1);
				}}
			/>
			{people.items.length > 0 && (
				<div className="space-y-1">
					<p className="text-[11px] font-medium text-sub">Filter by person</p>
					<div className="flex flex-wrap gap-2">
						{people.items.map((p) => {
							const active = personIds.includes(p.id);
							return (
								<button
									key={p.id}
									type="button"
									onClick={() =>
										change(setPersonIds)(
											active
												? personIds.filter((id) => id !== p.id)
												: [...personIds, p.id],
										)
									}
									className={`flex items-center gap-1.5 text-xs border rounded-full px-2 py-0.5 transition-colors ${active ? "border-accent bg-accent/10 text-accent-text" : "border-line hover:border-sub"}`}
								>
									<span className="size-4 rounded-full overflow-hidden shrink-0 bg-chip flex items-center justify-center text-[9px] font-medium text-chip-fg">
										{p.avatar_path ? (
											<img
												src={getAvatarUrl(p.id)}
												alt={p.name}
												className="size-full object-cover"
											/>
										) : (
											p.name.charAt(0).toUpperCase()
										)}
									</span>
									{p.name}
								</button>
							);
						})}
						{personIds.length > 0 && (
							<button
								type="button"
								onClick={() => change(setPersonIds)([])}
								className="text-xs text-sub hover:text-ink"
							>
								Clear
							</button>
						)}
					</div>
				</div>
			)}
		</div>
	);

	if (!data) return filterUi;

	if (data.items.length === 0) {
		return (
			<>
				{filterUi}
				<p className="text-sm text-sub">No notes found.</p>
			</>
		);
	}

	return (
		<div>
			{filterUi}
			<div className="space-y-8">
				{groupByMonth(data.items).map((group) => (
					<div key={group.label}>
						<h2 className="text-[11px] font-semibold uppercase tracking-widest text-sub mb-4 pb-2 border-b border-line-soft">
							{group.label}
						</h2>
						<ul className="divide-y divide-line-soft">
							{group.items.map((n) => {
								const label = n.person_nickname || n.person_name;
								return (
									<li
										key={n.id}
										className="flex flex-col sm:flex-row gap-2 sm:gap-4 py-4 first:pt-0"
									>
										<Link
											to="/people/$personId"
											params={{ personId: String(n.person_id) }}
											className="sm:w-44 shrink-0 flex items-center gap-2 hover:underline"
										>
											<span className="size-7 rounded-full overflow-hidden shrink-0 bg-chip flex items-center justify-center text-[11px] font-medium text-chip-fg font-mono">
												{n.person_has_avatar ? (
													<img
														src={getAvatarUrl(n.person_id)}
														alt={label}
														className="size-full object-cover"
													/>
												) : (
													<span>{label.charAt(0).toUpperCase()}</span>
												)}
											</span>
											<span className="min-w-0">
												<span className="block text-[13px] text-ink truncate">
													{label}
												</span>
												<span className="block font-mono text-[10px] text-sub">
													{formatDateTime(n.created_at)}
												</span>
											</span>
										</Link>
										<div className="flex-1 min-w-0">
											{n.title && (
												<p className="text-[14px] font-medium text-ink leading-snug">
													{n.title}
												</p>
											)}
											{n.content && (
												<p className="text-[12px] text-sub whitespace-pre-wrap leading-relaxed">
													{n.content}
												</p>
											)}
										</div>
									</li>
								);
							})}
						</ul>
					</div>
				))}
				{data.total > PAGE_SIZE && (
					<JournalPagination
						page={data.page}
						pageSize={PAGE_SIZE}
						total={data.total}
						onPageChange={setPage}
					/>
				)}
			</div>
		</div>
	);
}

export function AllNotesList() {
	return (
		<QueryBoundary>
			<AllNotesListInner />
		</QueryBoundary>
	);
}
