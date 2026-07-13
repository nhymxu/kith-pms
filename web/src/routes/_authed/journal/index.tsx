import {
	keepPreviousData,
	useQuery,
	useSuspenseQuery,
} from "@tanstack/react-query";
import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { z } from "zod";
import { PageSizeSelector } from "#/components/page-size-selector";
import { Button } from "#/components/ui/button";
import { listJournal } from "#/endpoints/journal";
import { listJournalLabels } from "#/endpoints/journal-labels";
import { listPeople } from "#/endpoints/people";
import { getSettings } from "#/endpoints/settings";
import { JournalPagination } from "#/features/journal/journal-pagination";
import { JournalTimeline } from "#/features/journal/journal-timeline";
import { usePageSizeOverride } from "#/lib/use-page-size-override";
import { keys } from "#/query-keys";

const searchSchema = z.object({
	q: z.string().optional(),
	page: z.coerce.number().min(1).optional().default(1),
	page_size: z.coerce.number().min(1).max(200).optional(),
	from_date: z.string().optional(),
	to_date: z.string().optional(),
	people: z.array(z.coerce.number()).optional(),
	labels: z.array(z.coerce.number()).optional(),
});

export const Route = createFileRoute("/_authed/journal/")({
	validateSearch: searchSchema,
	component: JournalPage,
	pendingComponent: () => (
		<p className="text-sm font-base text-foreground/60 py-4">Loading…</p>
	),
	errorComponent: () => (
		<p className="text-sm font-base text-destructive">
			Failed to load journal.
		</p>
	),
});

const PEOPLE_FILTER_KEY = "journal.filter.people_with_journal";

function JournalPage() {
	const navigate = useNavigate();
	const search = Route.useSearch();

	const [onlyWithJournal, setOnlyWithJournal] = useState<boolean>(
		() => localStorage.getItem(PEOPLE_FILTER_KEY) !== "false",
	);

	const { data: settingsData } = useQuery({
		queryKey: ["settings"],
		queryFn: getSettings,
	});
	const { override, setPageSize, clear } = usePageSizeOverride("journal");
	const effectivePageSize =
		search.page_size ?? override ?? settingsData?.default_page_size ?? 20;

	const { data } = useQuery({
		queryKey: keys.journal.list({
			page: search.page,
			page_size: effectivePageSize,
			person_ids: search.people,
			from_date: search.from_date,
			to_date: search.to_date,
		}),
		queryFn: () =>
			listJournal({
				q: search.q,
				page: search.page,
				page_size: effectivePageSize,
				from_date: search.from_date,
				to_date: search.to_date,
				person_ids: search.people,
				labels: search.labels,
			}),
		placeholderData: keepPreviousData,
	});

	const { data: allPeople } = useSuspenseQuery({
		queryKey: keys.people.list({
			page_size: 500,
			has_journal: onlyWithJournal || undefined,
		}),
		queryFn: () =>
			listPeople({ page_size: 500, has_journal: onlyWithJournal || undefined }),
	});

	const { data: allJournalLabels } = useSuspenseQuery({
		queryKey: keys.journalLabels.list(),
		queryFn: listJournalLabels,
	});

	return (
		<div className="space-y-4">
			<div className="flex items-center justify-between">
				<h1 className="text-[18px] font-semibold tracking-tight text-ink font-display">
					Journal
				</h1>
				<Button asChild>
					<Link to="/journal/new">New Entry</Link>
				</Button>
			</div>

			{/* Date range filter */}
			<div className="flex flex-wrap gap-3 items-end">
				<div className="space-y-1">
					<label
						htmlFor="journal-from-date"
						className="text-[11px] font-medium text-sub"
					>
						From
					</label>
					<input
						id="journal-from-date"
						type="date"
						value={search.from_date ?? ""}
						onChange={(e) =>
							void navigate({
								to: "/journal",
								search: {
									...search,
									from_date: e.target.value || undefined,
									page: 1,
								},
							})
						}
						className="h-9 border-bw border-line rounded-md bg-field px-3 text-[13px] focus:outline-none focus:ring-2 focus:ring-ring"
					/>
				</div>
				<div className="space-y-1">
					<label
						htmlFor="journal-to-date"
						className="text-[11px] font-medium text-sub"
					>
						To
					</label>
					<input
						id="journal-to-date"
						type="date"
						value={search.to_date ?? ""}
						onChange={(e) =>
							void navigate({
								to: "/journal",
								search: {
									...search,
									to_date: e.target.value || undefined,
									page: 1,
								},
							})
						}
						className="h-9 border-bw border-line rounded-md bg-field px-3 text-[13px] focus:outline-none focus:ring-2 focus:ring-ring"
					/>
				</div>
				{(search.from_date || search.to_date) && (
					<Button
						variant="neutral"
						size="sm"
						onClick={() =>
							void navigate({
								to: "/journal",
								search: {
									...search,
									from_date: undefined,
									to_date: undefined,
									page: 1,
								},
							})
						}
					>
						Clear dates
					</Button>
				)}
			</div>

			{/* People filter */}
			<div className="space-y-1">
				<div className="flex items-center justify-between">
					<p className="text-[11px] font-medium text-sub">Filter by person</p>
					<label className="flex items-center gap-1.5 cursor-pointer select-none">
						<input
							type="checkbox"
							checked={onlyWithJournal}
							onChange={(e) => {
								const val = e.target.checked;
								setOnlyWithJournal(val);
								localStorage.setItem(PEOPLE_FILTER_KEY, String(val));
							}}
							className="accent-accent size-3"
						/>
						<span className="text-[11px] text-sub">With journal only</span>
					</label>
				</div>
				{allPeople.items.length > 0 && (
					<div className="flex flex-wrap gap-2">
						{allPeople.items.map((p) => {
							const active = (search.people ?? []).includes(p.id);
							const base =
								(import.meta.env.VITE_API_BASE_URL as string | undefined) ?? "";
							return (
								<button
									key={p.id}
									type="button"
									onClick={() => {
										const next = active
											? (search.people ?? []).filter((id) => id !== p.id)
											: [...(search.people ?? []), p.id];
										void navigate({
											to: "/journal",
											search: {
												...search,
												people: next.length ? next : undefined,
												page: 1,
											},
										});
									}}
									className={`flex items-center gap-1.5 text-xs border rounded-full px-2 py-0.5 transition-colors ${active ? "border-accent bg-accent/10 text-accent-text" : "border-line hover:border-sub"}`}
								>
									<span className="size-4 rounded-full overflow-hidden shrink-0 bg-chip flex items-center justify-center text-[9px] font-medium text-chip-fg">
										{p.avatar_path ? (
											<img
												src={`${base}/v1/people/${p.id}/avatar`}
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
						{(search.people?.length ?? 0) > 0 && (
							<button
								type="button"
								onClick={() =>
									void navigate({
										to: "/journal",
										search: { ...search, people: undefined, page: 1 },
									})
								}
								className="text-xs text-sub hover:text-ink"
							>
								Clear
							</button>
						)}
					</div>
				)}
			</div>

			{/* Journal label filter */}
			{allJournalLabels.length > 0 && (
				<div className="space-y-1">
					<p className="text-[11px] font-medium text-sub">Filter by label</p>
					<div className="flex flex-wrap gap-2">
						{allJournalLabels.map((l) => {
							const active = (search.labels ?? []).includes(l.id);
							return (
								<button
									key={l.id}
									type="button"
									onClick={() => {
										const next = active
											? (search.labels ?? []).filter((id) => id !== l.id)
											: [...(search.labels ?? []), l.id];
										void navigate({
											to: "/journal",
											search: {
												...search,
												labels: next.length ? next : undefined,
												page: 1,
											},
										});
									}}
									className={`text-xs border rounded-md px-2 py-1 transition-colors ${active ? "border-main bg-main/10" : "border-line hover:border-sub"}`}
									style={active ? { borderColor: l.color } : undefined}
								>
									{l.name}
								</button>
							);
						})}
						{(search.labels?.length ?? 0) > 0 && (
							<button
								type="button"
								onClick={() =>
									void navigate({
										to: "/journal",
										search: { ...search, labels: undefined, page: 1 },
									})
								}
								className="text-xs text-sub hover:text-ink"
							>
								Clear
							</button>
						)}
					</div>
				</div>
			)}

			<JournalTimeline data={data?.items ?? []} />

			<div className="flex items-center justify-center gap-4 pt-2">
				{data && data.total > data.page_size && (
					<JournalPagination
						page={data.page}
						pageSize={data.page_size}
						total={data.total}
						onPageChange={(page) =>
							void navigate({ to: "/journal", search: { ...search, page } })
						}
					/>
				)}
				<PageSizeSelector
					value={effectivePageSize}
					hasOverride={override !== null}
					onChange={(n) => {
						setPageSize(n);
						void navigate({
							to: "/journal",
							search: { ...search, page_size: undefined, page: 1 },
						});
					}}
					onReset={() => {
						clear();
						void navigate({
							to: "/journal",
							search: { ...search, page_size: undefined, page: 1 },
						});
					}}
				/>
			</div>
		</div>
	);
}
