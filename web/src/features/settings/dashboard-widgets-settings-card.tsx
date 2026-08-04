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

const MIN_COUNT = 1;
const MAX_COUNT = 20;

export function DashboardWidgetsSettingsCard({
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
	const [favoritesCount, setFavoritesCount] = useState(5);
	const [lastContactCount, setLastContactCount] = useState(5);

	const [synced, setSynced] = useState(false);
	if (apiSettings && !isPlaceholderData && !synced) {
		setFavoritesCount(apiSettings.dashboard_favorites_count);
		setLastContactCount(apiSettings.dashboard_last_contact_count);
		setSynced(true);
	}

	const widgetsMutation = useMutation({
		mutationFn: () =>
			updateSettings(
				buildPayload({
					dashboard_favorites_count: favoritesCount,
					dashboard_last_contact_count: lastContactCount,
				}),
			),
		onSuccess: (updated) => {
			queryClient.setQueryData(["settings"], updated);
		},
	});

	function clamp(value: number): number {
		if (Number.isNaN(value)) return MIN_COUNT;
		return Math.min(MAX_COUNT, Math.max(MIN_COUNT, value));
	}

	return (
		<Card>
			<CardHeader>
				<CardTitle className="text-[14px] font-medium text-ink">
					Dashboard Widgets
				</CardTitle>
				<CardDescription className="text-[12px] text-sub">
					How many people to show in the Favorites and Last Contact widgets on
					the dashboard.
				</CardDescription>
			</CardHeader>
			<CardContent className="space-y-4">
				<div className="space-y-2">
					<Label className="text-[13px]">Favorites</Label>
					<input
						type="number"
						min={MIN_COUNT}
						max={MAX_COUNT}
						value={favoritesCount}
						onChange={(e) =>
							setFavoritesCount(clamp(parseInt(e.target.value, 10)))
						}
						className="h-9 w-24 border-field-bw border-field-line rounded-base bg-field px-3 text-[13px] focus:outline-none focus:ring-2 focus:ring-ring"
					/>
				</div>

				<div className="space-y-2">
					<Label className="text-[13px]">Last Contact</Label>
					<input
						type="number"
						min={MIN_COUNT}
						max={MAX_COUNT}
						value={lastContactCount}
						onChange={(e) =>
							setLastContactCount(clamp(parseInt(e.target.value, 10)))
						}
						className="h-9 w-24 border-field-bw border-field-line rounded-base bg-field px-3 text-[13px] focus:outline-none focus:ring-2 focus:ring-ring"
					/>
				</div>

				<Button
					onClick={() => widgetsMutation.mutate()}
					size="sm"
					disabled={widgetsMutation.isPending}
				>
					{widgetsMutation.isPending
						? "Saving…"
						: widgetsMutation.isSuccess
							? "Saved!"
							: "Save"}
				</Button>
				{widgetsMutation.isError && (
					<p className="text-[12px] text-danger-fg">
						Failed to save. Please try again.
					</p>
				)}
			</CardContent>
		</Card>
	);
}
