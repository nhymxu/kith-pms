import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { useCallback, useEffect, useRef } from "react";
import { z } from "zod";
import { Button } from "#/components/ui/button";
import { getSettings } from "#/endpoints/settings";
import { PeopleTable } from "#/features/people/people-table";
import { initialSearch } from "#/lib/initial-search";

const VALID_SORTS = [
	"name",
	"-name",
	"last_contact",
	"-last_contact",
	"-favorite",
] as const;
type SortValue = (typeof VALID_SORTS)[number];

const searchSchema = z.object({
	q: z.string().optional(),
	page: z.coerce.number().min(1).optional().default(1),
	page_size: z.coerce.number().min(1).max(100).optional().default(20),
	labels: z.array(z.coerce.number()).optional(),
	sort: z.enum(VALID_SORTS).optional().default("name"),
	favorite_only: z.coerce.boolean().optional(),
	favorite_first: z.coerce.boolean().optional(),
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
		favorite_only: searchFavoriteOnly,
		favorite_first: searchFavoriteFirst,
	} = search;

	const { data: settingsData } = useQuery({
		queryKey: ["settings"],
		queryFn: getSettings,
	});
	const appliedDefaultRef = useRef(false);

	// One-shot: apply settings-derived defaults only on a bare first visit (no
	// explicit sort/favorite params in the raw URL). Checking the URL captured
	// at app boot (not `window.location.search`, which the router normalizes
	// with schema defaults before this effect ever runs, and not the
	// zod-parsed `search`, which always looks "already set" because of
	// `.default("name")`) is the only way to see the URL as the user actually
	// navigated to it.
	useEffect(() => {
		if (appliedDefaultRef.current || !settingsData) return;
		appliedDefaultRef.current = true;

		const raw = new URLSearchParams(initialSearch);
		const hasExplicitParams =
			raw.has("sort") || raw.has("favorite_only") || raw.has("favorite_first");
		if (hasExplicitParams) return;

		void navigate({
			to: "/people",
			search: {
				...search,
				sort: settingsData.default_people_sort as SortValue,
				favorite_first: settingsData.favorite_first_default || undefined,
			},
			replace: true,
		});
	}, [settingsData, navigate, search]);

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
					favorite_only: searchFavoriteOnly,
					favorite_first: searchFavoriteFirst,
				},
			});
		},
		[
			searchLabels,
			searchPageSize,
			searchSort,
			searchFavoriteOnly,
			searchFavoriteFirst,
			navigate,
		],
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
					favorite_only: searchFavoriteOnly,
					favorite_first: searchFavoriteFirst,
				},
			});
		},
		[
			searchQ,
			searchPageSize,
			searchSort,
			searchFavoriteOnly,
			searchFavoriteFirst,
			navigate,
		],
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
					favorite_only: searchFavoriteOnly,
					favorite_first: searchFavoriteFirst,
				},
			});
		},
		[
			searchQ,
			searchLabels,
			searchPageSize,
			searchSort,
			searchFavoriteOnly,
			searchFavoriteFirst,
			navigate,
		],
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
					favorite_only: searchFavoriteOnly,
					favorite_first: searchFavoriteFirst,
				},
			});
		},
		[
			searchQ,
			searchLabels,
			searchPageSize,
			searchFavoriteOnly,
			searchFavoriteFirst,
			navigate,
		],
	);

	const handleFavoriteOnlyChange = useCallback(
		(favoriteOnly: boolean) => {
			void navigate({
				to: "/people",
				search: {
					q: searchQ || undefined,
					page: 1,
					page_size: searchPageSize,
					labels: searchLabels,
					sort: searchSort,
					favorite_only: favoriteOnly || undefined,
					favorite_first: searchFavoriteFirst,
				},
			});
		},
		[
			searchQ,
			searchLabels,
			searchPageSize,
			searchSort,
			searchFavoriteFirst,
			navigate,
		],
	);

	const handleFavoriteFirstChange = useCallback(
		(favoriteFirst: boolean) => {
			void navigate({
				to: "/people",
				search: {
					q: searchQ || undefined,
					page: 1,
					page_size: searchPageSize,
					labels: searchLabels,
					sort: searchSort,
					favorite_only: searchFavoriteOnly,
					favorite_first: favoriteFirst || undefined,
				},
			});
		},
		[
			searchQ,
			searchLabels,
			searchPageSize,
			searchSort,
			searchFavoriteOnly,
			navigate,
		],
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
				favoriteOnly={search.favorite_only}
				favoriteFirst={search.favorite_first}
				allowToggle={settingsData?.allow_favorite_toggle_on_list ?? true}
				onSearchChange={handleSearchChange}
				onLabelsChange={handleLabelsChange}
				onPageChange={handlePageChange}
				onSortChange={handleSortChange}
				onFavoriteOnlyChange={handleFavoriteOnlyChange}
				onFavoriteFirstChange={handleFavoriteFirstChange}
			/>
		</div>
	);
}
