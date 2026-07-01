import { useMutation, type useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { Button } from "#/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "#/components/ui/card";
import { Label } from "#/components/ui/label";
import { updateSettings } from "#/endpoints/settings";
import type { UserSettings } from "#/schemas/settings";

type DefaultPeopleSort = UserSettings["default_people_sort"];

const SORT_OPTIONS: { value: DefaultPeopleSort; label: string }[] = [
	{ value: "name", label: "Name A→Z" },
	{ value: "-name", label: "Name Z→A" },
	{ value: "-last_contact", label: "Last contact: newest" },
	{ value: "last_contact", label: "Last contact: oldest" },
];

export function FavoritesSettingsCard({
	apiSettings,
	isPlaceholderData,
	buildPayload,
	queryClient,
}: {
	apiSettings: UserSettings | undefined;
	isPlaceholderData: boolean;
	buildPayload: (
		overrides?: Partial<Parameters<typeof updateSettings>[0]>,
	) => Parameters<typeof updateSettings>[0];
	queryClient: ReturnType<typeof useQueryClient>;
}) {
	const [favoritesDefaults, setFavoritesDefaults] = useState<{
		allowToggleOnList: boolean;
		favoriteFirstDefault: boolean;
		defaultSort: DefaultPeopleSort;
	}>(() => ({
		allowToggleOnList: true,
		favoriteFirstDefault: false,
		defaultSort: "name",
	}));

	const [synced, setSynced] = useState(false);
	if (apiSettings && !isPlaceholderData && !synced) {
		setFavoritesDefaults({
			allowToggleOnList: apiSettings.allow_favorite_toggle_on_list,
			favoriteFirstDefault: apiSettings.favorite_first_default,
			defaultSort: apiSettings.default_people_sort,
		});
		setSynced(true);
	}

	const favoritesMutation = useMutation({
		mutationFn: () =>
			updateSettings(
				buildPayload({
					allow_favorite_toggle_on_list: favoritesDefaults.allowToggleOnList,
					favorite_first_default: favoritesDefaults.favoriteFirstDefault,
					default_people_sort: favoritesDefaults.defaultSort,
				}),
			),
		onSuccess: (updated) => {
			queryClient.setQueryData(["settings"], updated);
		},
	});

	return (
		<Card>
			<CardHeader>
				<CardTitle className="text-[14px] font-medium text-zinc-900">
					Favorites
				</CardTitle>
				<CardDescription className="text-[12px] text-zinc-500">
					Controls how favoriting behaves on the People list. Toggling favorite
					status on a person's own page is always available.
				</CardDescription>
			</CardHeader>
			<CardContent className="space-y-4">
				<label className="flex items-center gap-3 cursor-pointer">
					<input
						type="checkbox"
						checked={favoritesDefaults.allowToggleOnList}
						onChange={(e) =>
							setFavoritesDefaults((f) => ({
								...f,
								allowToggleOnList: e.target.checked,
							}))
						}
						className="accent-indigo-600"
					/>
					<span className="text-[13px] text-zinc-700">
						Allow favorite/unfavorite directly on the People list
					</span>
				</label>

				<label className="flex items-center gap-3 cursor-pointer">
					<input
						type="checkbox"
						checked={favoritesDefaults.favoriteFirstDefault}
						onChange={(e) =>
							setFavoritesDefaults((f) => ({
								...f,
								favoriteFirstDefault: e.target.checked,
							}))
						}
						className="accent-indigo-600"
					/>
					<span className="text-[13px] text-zinc-700">
						Show favorites first by default
					</span>
				</label>

				<div className="space-y-2">
					<Label className="text-[13px]">Default sort</Label>
					<div className="space-y-1.5">
						{SORT_OPTIONS.map((opt) => (
							<label
								key={opt.value}
								className="flex items-center gap-3 cursor-pointer"
							>
								<input
									type="radio"
									name="defaultPeopleSort"
									value={opt.value}
									checked={favoritesDefaults.defaultSort === opt.value}
									onChange={() =>
										setFavoritesDefaults((f) => ({
											...f,
											defaultSort: opt.value,
										}))
									}
									className="accent-indigo-600"
								/>
								<span className="text-[13px] text-zinc-700">{opt.label}</span>
							</label>
						))}
					</div>
				</div>

				<Button
					onClick={() => favoritesMutation.mutate()}
					size="sm"
					disabled={favoritesMutation.isPending}
				>
					{favoritesMutation.isPending
						? "Saving…"
						: favoritesMutation.isSuccess
							? "Saved!"
							: "Save defaults"}
				</Button>
				{favoritesMutation.isError && (
					<p className="text-[12px] text-red-500">
						Failed to save. Please try again.
					</p>
				)}
			</CardContent>
		</Card>
	);
}
