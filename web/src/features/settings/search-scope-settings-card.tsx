import { useMutation, type useQueryClient } from "@tanstack/react-query";
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
import { updateSettings } from "#/endpoints/settings";
import type { UserSettings } from "#/schemas/settings";

type SearchScopeToken = UserSettings["search_scope"][number];

const ALL_SCOPES: SearchScopeToken[] = ["people", "journal", "gifts", "notes"];

export function SearchScopeSettingsCard({
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
	const [searchScope, setSearchScope] =
		useState<SearchScopeToken[]>(ALL_SCOPES);

	const [synced, setSynced] = useState(false);
	if (apiSettings && !isPlaceholderData && !synced) {
		setSearchScope(apiSettings.search_scope);
		setSynced(true);
	}

	const searchScopeMutation = useMutation({
		mutationFn: () =>
			updateSettings(buildPayload({ search_scope: searchScope })),
		onSuccess: (updated) => {
			queryClient.setQueryData(["settings"], updated);
		},
	});

	function toggle(value: SearchScopeToken, checked: boolean) {
		setSearchScope((prev) =>
			checked ? [...prev, value] : prev.filter((v) => v !== value),
		);
	}

	return (
		<Card>
			<CardHeader>
				<CardTitle className="text-[14px] font-medium text-ink">
					{t("general.searchScope.title")}
				</CardTitle>
				<CardDescription className="text-[12px] text-sub">
					{t("general.searchScope.description")}
				</CardDescription>
			</CardHeader>
			<CardContent className="space-y-4">
				<div className="space-y-1.5">
					{ALL_SCOPES.map((value) => (
						<label
							key={value}
							className="flex items-center gap-3 cursor-pointer"
						>
							<input
								type="checkbox"
								checked={searchScope.includes(value)}
								onChange={(e) => toggle(value, e.target.checked)}
								className="accent-accent"
							/>
							<span className="text-[13px] text-ink">
								{t(`general.searchScope.${value}`)}
							</span>
						</label>
					))}
				</div>

				<Button
					onClick={() => searchScopeMutation.mutate()}
					size="sm"
					disabled={searchScope.length === 0 || searchScopeMutation.isPending}
				>
					{searchScopeMutation.isPending
						? t("common.saving")
						: searchScopeMutation.isSuccess
							? t("common.saved")
							: t("general.searchScope.save")}
				</Button>
				{searchScope.length === 0 && (
					<p className="text-[12px] text-danger-fg">
						{t("general.searchScope.pickOne")}
					</p>
				)}
				{searchScopeMutation.isError && (
					<p className="text-[12px] text-danger-fg">{t("common.saveFailed")}</p>
				)}
			</CardContent>
		</Card>
	);
}
