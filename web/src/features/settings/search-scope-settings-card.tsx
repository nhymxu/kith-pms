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
import { updateSettings } from "#/endpoints/settings";
import type { UserSettings } from "#/schemas/settings";

type SearchScopeToken = UserSettings["search_scope"][number];

const SEARCH_SCOPE_OPTIONS: { value: SearchScopeToken; label: string }[] = [
	{ value: "people", label: "People" },
	{ value: "journal", label: "Journal" },
	{ value: "gifts", label: "Gifts" },
	{ value: "notes", label: "Notes" },
];

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
					Search scope
				</CardTitle>
				<CardDescription className="text-[12px] text-sub">
					Choose what the search box returns. Indexes always stay up to date —
					unchecking a type only hides it from results.
				</CardDescription>
			</CardHeader>
			<CardContent className="space-y-4">
				<div className="space-y-1.5">
					{SEARCH_SCOPE_OPTIONS.map((opt) => (
						<label
							key={opt.value}
							className="flex items-center gap-3 cursor-pointer"
						>
							<input
								type="checkbox"
								checked={searchScope.includes(opt.value)}
								onChange={(e) => toggle(opt.value, e.target.checked)}
								className="accent-accent"
							/>
							<span className="text-[13px] text-ink">{opt.label}</span>
						</label>
					))}
				</div>

				<Button
					onClick={() => searchScopeMutation.mutate()}
					size="sm"
					disabled={searchScope.length === 0 || searchScopeMutation.isPending}
				>
					{searchScopeMutation.isPending
						? "Saving…"
						: searchScopeMutation.isSuccess
							? "Saved!"
							: "Save scope"}
				</Button>
				{searchScope.length === 0 && (
					<p className="text-[12px] text-danger-fg">Pick at least one type.</p>
				)}
				{searchScopeMutation.isError && (
					<p className="text-[12px] text-danger-fg">
						Failed to save. Please try again.
					</p>
				)}
			</CardContent>
		</Card>
	);
}
