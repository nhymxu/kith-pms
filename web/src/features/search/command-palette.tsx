import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import {
	CommandDialog,
	CommandEmpty,
	CommandGroup,
	CommandInput,
	CommandItem,
	CommandList,
} from "#/components/ui/command";
import { searchAll } from "#/endpoints/search";
import { getSettings } from "#/endpoints/settings";
import { useDebounce } from "#/hooks/use-debounce";
import { keys } from "#/query-keys";
import type { SearchItem } from "#/schemas/search";

type ResultGroup = "people" | "journal" | "gifts" | "notes";

const GROUP_LABELS: Record<ResultGroup, string> = {
	people: "People",
	journal: "Journal",
	gifts: "Gifts",
	notes: "Notes",
};

export function CommandPalette() {
	const [open, setOpen] = useState(false);
	const [query, setQuery] = useState("");
	const debouncedQuery = useDebounce(query, 250);
	const navigate = useNavigate();

	useEffect(() => {
		function onOpenEvent() {
			setOpen(true);
		}

		function onKeyDown(e: KeyboardEvent) {
			const isCmdK = (e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k";
			if (!isCmdK) return;
			e.preventDefault();
			setOpen((prev) => !prev);
		}

		window.addEventListener("kith:open-command-palette", onOpenEvent);
		window.addEventListener("keydown", onKeyDown);
		return () => {
			window.removeEventListener("kith:open-command-palette", onOpenEvent);
			window.removeEventListener("keydown", onKeyDown);
		};
	}, []);

	useEffect(() => {
		if (!open) setQuery("");
	}, [open]);

	const trimmedQuery = debouncedQuery.trim();

	// Settings query is shared with the settings page cache (["settings"]); if
	// not yet loaded, `types` is omitted and the backend defaults to all groups.
	const { data: settingsData } = useQuery({
		queryKey: ["settings"],
		queryFn: getSettings,
		enabled: open,
	});
	const searchScope = settingsData?.search_scope;

	const { data, isFetching } = useQuery({
		queryKey: keys.search.query(trimmedQuery, searchScope),
		queryFn: () => searchAll(trimmedQuery, searchScope),
		enabled: open && trimmedQuery.length > 0,
		staleTime: 10_000,
	});

	// Navigates by typed route + id rather than the raw `url` field from the API —
	// keeps navigation targets confined to known routes (see phase security note).
	function handleSelect(group: ResultGroup, item: SearchItem) {
		setOpen(false);
		const id = String(item.id);
		switch (group) {
			case "people":
				navigate({ to: "/people/$personId", params: { personId: id } });
				break;
			case "journal":
				navigate({ to: "/journal/$entryId", params: { entryId: id } });
				break;
			case "gifts":
				navigate({ to: "/gifts/$giftId", params: { giftId: id } });
				break;
			case "notes":
				if (item.url === "/notes") {
					navigate({ to: "/notes" });
				} else {
					navigate({
						to: "/people/$personId",
						params: { personId: item.url.replace("/people/", "") },
					});
				}
				break;
		}
	}

	const groups: Array<{ key: ResultGroup; items: SearchItem[] }> = data
		? [
				{ key: "people", items: data.people },
				{ key: "journal", items: data.journal },
				{ key: "gifts", items: data.gifts },
				{ key: "notes", items: data.notes },
			]
		: [];
	const hasResults = groups.some((g) => g.items.length > 0);

	return (
		<CommandDialog
			open={open}
			onOpenChange={setOpen}
			title="Search"
			description="Search people, journal entries, gifts, and notes"
			shouldFilter={false}
		>
			<CommandInput
				value={query}
				onValueChange={setQuery}
				placeholder="Search people, notes, gifts…"
			/>
			<CommandList>
				{trimmedQuery.length === 0 && (
					<CommandEmpty>Type to search…</CommandEmpty>
				)}
				{trimmedQuery.length > 0 && !isFetching && !hasResults && (
					<CommandEmpty>No results found.</CommandEmpty>
				)}
				{groups.map(
					(group) =>
						group.items.length > 0 && (
							<CommandGroup key={group.key} heading={GROUP_LABELS[group.key]}>
								{group.items.map((item) => (
									<CommandItem
										key={`${group.key}-${item.id}`}
										value={`${group.key}-${item.id}-${item.title}`}
										onSelect={() => handleSelect(group.key, item)}
									>
										<div className="flex min-w-0 flex-1 flex-col">
											<span className="truncate">{item.title}</span>
											{item.subtitle && (
												<span className="truncate text-xs text-sub">
													{item.subtitle}
												</span>
											)}
										</div>
									</CommandItem>
								))}
							</CommandGroup>
						),
				)}
			</CommandList>
		</CommandDialog>
	);
}
