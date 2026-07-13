import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Check } from "lucide-react";
import { useEffect, useState } from "react";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "#/components/ui/card";
import { getSettings, updateSettings } from "#/endpoints/settings";
import { getUserPrefs } from "#/lib/format-datetime";
import {
	getTheme,
	setTheme,
	THEME_META,
	type Theme,
	type ThemeMeta,
} from "#/lib/theme";
import type { UserSettings } from "#/schemas/settings";

export const Route = createFileRoute("/_authed/settings/_layout/appearance")({
	component: AppearancePage,
});

function ThemeSwatch({ meta }: { meta: ThemeMeta }) {
	return (
		<div
			className="flex h-10 w-full items-center gap-1.5 rounded-base border border-line px-2"
			style={{ background: meta.swatch.bg }}
			aria-hidden="true"
		>
			<span
				className="h-6 flex-1 rounded-sm border border-black/10"
				style={{ background: meta.swatch.panel }}
			/>
			<span
				className="size-6 shrink-0 rounded-full"
				style={{ background: meta.swatch.accent }}
			/>
			<span
				className="h-6 w-3 shrink-0 rounded-sm"
				style={{ background: meta.swatch.ink }}
			/>
		</div>
	);
}

function AppearancePage() {
	const queryClient = useQueryClient();

	const { data: apiSettings, isPlaceholderData } = useQuery({
		queryKey: ["settings"],
		queryFn: getSettings,
		placeholderData: (): UserSettings => {
			const p = getUserPrefs();
			return {
				date_format: p.dateFormat,
				time_format: p.timeFormat,
				timezone: p.timezone,
				theme: p.theme,
				audit_log_retention_days: 0,
				network_color_by: p.networkColorBy,
				network_show_avatar: p.networkShowAvatar,
				network_show_only_mine: p.networkShowOnlyMine,
				network_show_unconnected: p.networkShowUnconnected,
				network_only_mine_depth: p.networkOnlyMineDepth,
				allow_favorite_toggle_on_list: true,
				favorite_first_default: false,
				default_people_sort: "name",
				default_page_size: 25,
				dashboard_favorites_count: 5,
				dashboard_last_contact_count: 5,
			};
		},
	});

	const [active, setActive] = useState<Theme>(() => getTheme());

	// Reflect the DB value once real settings arrive (cross-device change).
	useEffect(() => {
		if (apiSettings && !isPlaceholderData) {
			setActive(apiSettings.theme);
			setTheme(apiSettings.theme);
		}
	}, [apiSettings, isPlaceholderData]);

	const mutation = useMutation({
		mutationFn: (theme: Theme) => {
			if (!apiSettings) return Promise.reject(new Error("settings not loaded"));
			return updateSettings({ ...apiSettings, theme });
		},
		onSuccess: (saved) => {
			queryClient.setQueryData(["settings"], saved);
		},
	});

	const onSelect = (theme: Theme) => {
		if (theme === active) return;
		const prev = active;
		setActive(theme);
		setTheme(theme);
		mutation.mutate(theme, {
			onError: () => {
				setActive(prev);
				setTheme(prev);
			},
		});
	};

	return (
		<div className="space-y-6 max-w-3xl">
			<Card>
				<CardHeader>
					<CardTitle>Appearance</CardTitle>
					<CardDescription>
						Choose a theme. It applies instantly and is saved to your account.
					</CardDescription>
				</CardHeader>
				<CardContent>
					<div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
						{THEME_META.map((meta) => {
							const selected = active === meta.id;
							return (
								<button
									key={meta.id}
									type="button"
									aria-pressed={selected}
									disabled={isPlaceholderData || mutation.isPending}
									onClick={() => onSelect(meta.id)}
									className={`text-left rounded-base border-bw p-3 space-y-2 transition-colors disabled:opacity-60 ${
										selected
											? "border-accent ring-2 ring-ring"
											: "border-line hover:border-sub"
									}`}
								>
									<ThemeSwatch meta={meta} />
									<div className="flex items-center gap-1.5">
										<span className="text-[13px] font-medium text-ink">
											{meta.label}
										</span>
										{selected && (
											<Check className="size-3.5 text-accent-text shrink-0" />
										)}
									</div>
									<p className="text-[12px] text-sub">{meta.description}</p>
								</button>
							);
						})}
					</div>
					{mutation.isError && (
						<p className="text-[12px] text-danger-fg mt-3">
							Failed to save theme. Your previous theme was restored.
						</p>
					)}
				</CardContent>
			</Card>
		</div>
	);
}
