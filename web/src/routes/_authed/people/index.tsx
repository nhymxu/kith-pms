import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { useCallback } from "react";
import { z } from "zod";
import { Button } from "#/components/ui/button";
import { PeopleTable } from "#/features/people/people-table";

const searchSchema = z.object({
	q: z.string().optional(),
	page: z.coerce.number().min(1).optional().default(1),
	page_size: z.coerce.number().min(1).max(100).optional().default(20),
	labels: z.array(z.coerce.number()).optional(),
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
				},
			});
		},
		[searchLabels, searchPageSize, navigate],
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
				},
			});
		},
		[searchQ, searchPageSize, navigate],
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
				onSearchChange={handleSearchChange}
				onLabelsChange={handleLabelsChange}
				onPageChange={handlePageChange}
			/>
		</div>
	);
}
