import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { useCallback } from "react";
import { z } from "zod";
import { Button } from "#/components/ui/button";
import { PeopleTable } from "#/features/people/people-table";

const VALID_SORTS = ["name", "-name", "last_contact", "-last_contact"] as const;
type SortValue = (typeof VALID_SORTS)[number];

const searchSchema = z.object({
	q: z.string().optional(),
	page: z.coerce.number().min(1).optional().default(1),
	page_size: z.coerce.number().min(1).max(100).optional().default(20),
	labels: z.array(z.coerce.number()).optional(),
	sort: z.enum(VALID_SORTS).optional().default("name"),
});

export const Route = createFileRoute("/_authed/people/")({
	validateSearch: searchSchema,
	component: PeoplePage,
});

function PeoplePage() {
	const navigate = useNavigate();
	const search = Route.useSearch();
	const {
		q: searchQ,
		labels: searchLabels,
		page_size: searchPageSize,
		sort: searchSort,
	} = search;

	// Each handler only captures the fields it reads — excludes `page` so identity
	// stays stable when the page number changes and avoids triggering the debounced
	// search effect in PeopleTable on every page flip.
	const handleSearchChange = useCallback(
		(q: string) => {
			void navigate({
				to: "/people",
				search: {
					q: q || undefined,
					page: 1,
					page_size: searchPageSize,
					labels: searchLabels,
					sort: searchSort,
				},
			});
		},
		[searchLabels, searchPageSize, searchSort, navigate],
	);

	const handleLabelsChange = useCallback(
		(labels: number[]) => {
			void navigate({
				to: "/people",
				search: {
					q: searchQ || undefined,
					page: 1,
					page_size: searchPageSize,
					labels: labels.length ? labels : undefined,
					sort: searchSort,
				},
			});
		},
		[searchQ, searchPageSize, searchSort, navigate],
	);

	const handlePageChange = useCallback(
		(page: number) => {
			void navigate({
				to: "/people",
				search: {
					q: searchQ || undefined,
					page,
					page_size: searchPageSize,
					labels: searchLabels,
					sort: searchSort,
				},
			});
		},
		[searchQ, searchLabels, searchPageSize, searchSort, navigate],
	);

	const handleSortChange = useCallback(
		(sort: SortValue) => {
			void navigate({
				to: "/people",
				search: {
					q: searchQ || undefined,
					page: 1,
					page_size: searchPageSize,
					labels: searchLabels,
					sort,
				},
			});
		},
		[searchQ, searchLabels, searchPageSize, navigate],
	);

	return (
		<div className="space-y-4">
			<div className="flex items-center justify-between">
				<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">
					People
				</h1>
				<Button asChild>
					<Link to="/people/new">New Person</Link>
				</Button>
			</div>
			<PeopleTable
				q={search.q}
				labels={search.labels}
				page={search.page}
				page_size={search.page_size}
				sort={search.sort}
				onSearchChange={handleSearchChange}
				onLabelsChange={handleLabelsChange}
				onPageChange={handlePageChange}
				onSortChange={handleSortChange}
			/>
		</div>
	);
}
