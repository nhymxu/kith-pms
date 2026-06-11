import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { z } from "zod";
import { Button } from "#/components/ui/button";
import { listJournal } from "#/endpoints/journal";
import { listJournalLabels } from "#/endpoints/journal-labels";
import { listPeople } from "#/endpoints/people";
import { JournalTimeline } from "#/features/journal/journal-timeline";
import { keys } from "#/query-keys";

const searchSchema = z.object({
	q: z.string().optional(),
	page: z.coerce.number().min(1).optional().default(1),
	page_size: z.coerce.number().min(1).max(100).optional().default(20),
	from_date: z.string().optional(),
	to_date: z.string().optional(),
	people: z.array(z.coerce.number()).optional(),
	labels: z.array(z.coerce.number()).optional(),
});

export const Route = createFileRoute("/_authed/journal/")({
	validateSearch: searchSchema,
	component: JournalPage,
});

function JournalPage() {
	const navigate = useNavigate();
	const search = Route.useSearch();

	const { data, isPending, isError, error } = useQuery({
		queryKey: keys.journal.list({
			page: search.page,
			page_size: search.page_size,
			person_ids: search.people,
			from_date: search.from_date,
			to_date: search.to_date,
		}),
		queryFn: () =>
			listJournal({
				q: search.q,
				page: search.page,
				page_size: search.page_size,
				from_date: search.from_date,
				to_date: search.to_date,
				person_ids: search.people,
				labels: search.labels,
			}),
	});

	const { data: allPeople } = useQuery({
		queryKey: keys.people.list({ page_size: 200 }),
		queryFn: () => listPeople({ page_size: 200 }),
	});

	const { data: allJournalLabels } = useQuery({
		queryKey: keys.journalLabels.list(),
		queryFn: listJournalLabels,
	});

	if (isError) {
		console.error("[journal] load error:", error);
		return (
			<p className="text-sm font-base text-destructive">
				Failed to load journal.
			</p>
		);
	}

	return (
		<div className="space-y-4">
			<div className="flex items-center justify-between">
				<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">
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
						className="text-[11px] font-medium text-zinc-500"
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
						className="h-9 border border-zinc-200 rounded-md bg-white px-3 text-[13px] focus:outline-none focus:ring-2 focus:ring-indigo-600"
					/>
				</div>
				<div className="space-y-1">
					<label
						htmlFor="journal-to-date"
						className="text-[11px] font-medium text-zinc-500"
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
						className="h-9 border border-zinc-200 rounded-md bg-white px-3 text-[13px] focus:outline-none focus:ring-2 focus:ring-indigo-600"
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
			{allPeople?.items && allPeople.items.length > 0 && (
				<div className="space-y-1">
					<p className="text-[11px] font-medium text-zinc-500">
						Filter by person
					</p>
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
									className={`flex items-center gap-1.5 text-xs border rounded-full px-2 py-0.5 transition-colors ${active ? "border-indigo-500 bg-indigo-50 text-indigo-700" : "border-zinc-200 hover:border-zinc-400"}`}
								>
									<span className="size-4 rounded-full overflow-hidden shrink-0 bg-zinc-100 flex items-center justify-center text-[9px] font-medium text-zinc-600">
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
								className="text-xs text-zinc-400 hover:text-zinc-700"
							>
								Clear
							</button>
						)}
					</div>
				</div>
			)}

			{/* Journal label filter */}
			{allJournalLabels && allJournalLabels.length > 0 && (
				<div className="space-y-1">
					<p className="text-[11px] font-medium text-zinc-500">
						Filter by label
					</p>
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
									className={`text-xs border rounded-md px-2 py-1 transition-colors ${active ? "border-main bg-main/10" : "border-zinc-200 hover:border-zinc-400"}`}
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
								className="text-xs text-zinc-400 hover:text-zinc-700"
							>
								Clear
							</button>
						)}
					</div>
				</div>
			)}

			{isPending ? (
				<p className="text-sm font-base text-foreground/60 py-4">Loading…</p>
			) : (
				<JournalTimeline data={data?.items ?? []} />
			)}
		</div>
	);
}
