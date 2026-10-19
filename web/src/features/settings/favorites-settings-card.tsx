import { useMutation, type useQueryClient } from "@tanstack/react-query";
import type { ParseKeys } from "i18next";
import { useState } from "react";
import { useTranslation } from "react-i18next";
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

type SettingsKey = ParseKeys<"settings">;

type DefaultPeopleSort = UserSettings["default_people_sort"];

const SORT_OPTIONS: { value: DefaultPeopleSort; labelKey: SettingsKey }[] = [
	{ value: "name", labelKey: "general.favorites.sort.nameAsc" },
	{ value: "-name", labelKey: "general.favorites.sort.nameDesc" },
	{
		value: "-last_contact",
		labelKey: "general.favorites.sort.lastContactNewest",
	},
	{
		value: "last_contact",
		labelKey: "general.favorites.sort.lastContactOldest",
	},
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
	const { t } = useTranslation("settings");
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
				<CardTitle className="text-[14px] font-medium text-ink">
					{t("general.favorites.title")}
				</CardTitle>
				<CardDescription className="text-[12px] text-sub">
					{t("general.favorites.description")}
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
						className="accent-accent"
					/>
					<span className="text-[13px] text-ink">
						{t("general.favorites.allowToggle")}
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
						className="accent-accent"
					/>
					<span className="text-[13px] text-ink">
						{t("general.favorites.favoriteFirst")}
					</span>
				</label>

				<div className="space-y-2">
					<Label className="text-[13px]">
						{t("general.favorites.defaultSort")}
					</Label>
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
									className="accent-accent"
								/>
								<span className="text-[13px] text-ink">{t(opt.labelKey)}</span>
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
						? t("common.saving")
						: favoritesMutation.isSuccess
							? t("common.saved")
							: t("general.favorites.save")}
				</Button>
				{favoritesMutation.isError && (
					<p className="text-[12px] text-danger-fg">{t("common.saveFailed")}</p>
				)}
			</CardContent>
		</Card>
	);
}
