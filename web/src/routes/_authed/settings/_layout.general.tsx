import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
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
import { runAuditCleanup } from "#/endpoints/audit";
import { getSettings, updateSettings } from "#/endpoints/settings";
import { DashboardWidgetsSettingsCard } from "#/features/settings/dashboard-widgets-settings-card";
import { FavoritesSettingsCard } from "#/features/settings/favorites-settings-card";
import { LanguageSettingsCard } from "#/features/settings/language-settings-card";
import { PageSizeSettingsCard } from "#/features/settings/page-size-settings-card";
import { RelationshipDisplaySettingsCard } from "#/features/settings/relationship-display-settings-card";
import { SearchScopeSettingsCard } from "#/features/settings/search-scope-settings-card";
import {
	type DateFormat,
	getNetworkPrefs,
	getUserPrefs,
	type NetworkColorBy,
	type NetworkOnlyMineDepth,
	type NumberFormat,
	saveUserPrefs,
	type TimeFormat,
} from "#/lib/format-datetime";
import { getTheme } from "#/lib/theme";
import type { SettingsResponse } from "#/schemas/settings";

type SettingsKey = ParseKeys<"settings">;

export const Route = createFileRoute("/_authed/settings/_layout/general")({
	component: GeneralSettingsPage,
});

const DATE_FORMAT_OPTIONS: {
	value: DateFormat;
	labelKey: SettingsKey;
	example: string;
}[] = [
	{
		value: "YYYY-MM-DD",
		labelKey: "general.dateTime.dateOptions.iso",
		example: "2026-05-19",
	},
	{
		value: "MM/DD/YYYY",
		labelKey: "general.dateTime.dateOptions.us",
		example: "05/19/2026",
	},
	{
		value: "DD/MM/YYYY",
		labelKey: "general.dateTime.dateOptions.european",
		example: "19/05/2026",
	},
];

const TIME_FORMAT_OPTIONS: {
	value: TimeFormat;
	labelKey: SettingsKey;
	example: string;
}[] = [
	{
		value: "24h",
		labelKey: "general.dateTime.timeOptions.h24",
		example: "14:30",
	},
	{
		value: "12h",
		labelKey: "general.dateTime.timeOptions.h12",
		example: "2:30 PM",
	},
];

const NUMBER_FORMAT_OPTIONS: {
	value: NumberFormat;
	labelKey: SettingsKey;
	example: string;
}[] = [
	{
		value: "1,234.56",
		labelKey: "general.dateTime.numberOptions.commaThousands",
		example: "1,234.56",
	},
	{
		value: "1.234,56",
		labelKey: "general.dateTime.numberOptions.dotThousands",
		example: "1.234,56",
	},
];

const COMMON_TIMEZONES = [
	"UTC",
	"America/New_York",
	"America/Chicago",
	"America/Denver",
	"America/Los_Angeles",
	"America/Sao_Paulo",
	"Europe/London",
	"Europe/Paris",
	"Europe/Berlin",
	"Europe/Moscow",
	"Asia/Dubai",
	"Asia/Kolkata",
	"Asia/Bangkok",
	"Asia/Saigon",
	"Asia/Singapore",
	"Asia/Shanghai",
	"Asia/Tokyo",
	"Asia/Seoul",
	"Australia/Sydney",
	"Pacific/Auckland",
];

function GeneralSettingsPage() {
	const { t } = useTranslation("settings");
	const queryClient = useQueryClient();

	const { data: apiSettings, isPlaceholderData } = useQuery({
		queryKey: ["settings"],
		queryFn: getSettings,
		// Seed local state from localStorage while the query loads.
		placeholderData: (): SettingsResponse => {
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
				nav_layout: p.navLayout,
				number_format: p.numberFormat,
				relationship_display: p.relationshipDisplay,
				language: p.language,
				search_scope: ["people", "journal", "gifts", "notes"],
				max_upload_size_mb: 32,
				image_max_edge_px: 1600,
				image_jpeg_quality: 85,
				people_delete_retention_days: 30,
			};
		},
	});

	const [prefs, setPrefs] = useState<{
		dateFormat: DateFormat;
		timeFormat: TimeFormat;
		timezone: string;
		numberFormat: NumberFormat;
	}>(() => {
		if (apiSettings) {
			return {
				dateFormat: apiSettings.date_format as DateFormat,
				timeFormat: apiSettings.time_format as TimeFormat,
				timezone: apiSettings.timezone,
				numberFormat: apiSettings.number_format as NumberFormat,
			};
		}
		return getUserPrefs();
	});

	const [retentionDays, setRetentionDays] = useState<number>(
		apiSettings?.audit_log_retention_days ?? 0,
	);

	const [networkDefaults, setNetworkDefaults] = useState<{
		colorBy: NetworkColorBy;
		showAvatar: boolean;
		showOnlyMine: boolean;
		showUnconnected: boolean;
		onlyMineDepth: NetworkOnlyMineDepth;
	}>(() => {
		const n = getNetworkPrefs();
		return {
			colorBy: n.networkColorBy,
			showAvatar: n.networkShowAvatar,
			showOnlyMine: n.networkShowOnlyMine,
			showUnconnected: n.networkShowUnconnected,
			onlyMineDepth: n.networkOnlyMineDepth,
		};
	});

	// Keep local form state in sync when the real query resolves (skip placeholder).
	const [synced, setSynced] = useState(false);
	if (apiSettings && !isPlaceholderData && !synced) {
		setPrefs({
			dateFormat: apiSettings.date_format as DateFormat,
			timeFormat: apiSettings.time_format as TimeFormat,
			timezone: apiSettings.timezone,
			numberFormat: apiSettings.number_format as NumberFormat,
		});
		setRetentionDays(apiSettings.audit_log_retention_days ?? 0);
		setNetworkDefaults({
			colorBy: apiSettings.network_color_by,
			showAvatar: apiSettings.network_show_avatar,
			showOnlyMine: apiSettings.network_show_only_mine,
			showUnconnected: apiSettings.network_show_unconnected,
			onlyMineDepth: (apiSettings.network_only_mine_depth ??
				"direct") as NetworkOnlyMineDepth,
		});
		setSynced(true);
	}

	const buildPayload = (
		overrides?: Partial<Parameters<typeof updateSettings>[0]>,
	) => ({
		date_format: prefs.dateFormat,
		time_format: prefs.timeFormat,
		timezone: prefs.timezone,
		number_format: prefs.numberFormat,
		theme: apiSettings?.theme ?? getTheme(),
		audit_log_retention_days:
			apiSettings?.audit_log_retention_days ?? retentionDays,
		network_color_by: networkDefaults.colorBy,
		network_show_avatar: networkDefaults.showAvatar,
		network_show_only_mine: networkDefaults.showOnlyMine,
		network_show_unconnected: networkDefaults.showUnconnected,
		network_only_mine_depth: networkDefaults.onlyMineDepth,
		allow_favorite_toggle_on_list:
			apiSettings?.allow_favorite_toggle_on_list ?? true,
		favorite_first_default: apiSettings?.favorite_first_default ?? false,
		default_people_sort: apiSettings?.default_people_sort ?? "name",
		default_page_size: apiSettings?.default_page_size ?? 25,
		dashboard_favorites_count: apiSettings?.dashboard_favorites_count ?? 5,
		dashboard_last_contact_count:
			apiSettings?.dashboard_last_contact_count ?? 5,
		nav_layout: apiSettings?.nav_layout ?? getUserPrefs().navLayout,
		relationship_display:
			apiSettings?.relationship_display ?? getUserPrefs().relationshipDisplay,
		language: apiSettings?.language ?? getUserPrefs().language,
		search_scope: apiSettings?.search_scope ?? [
			"people",
			"journal",
			"gifts",
			"notes",
		],
		...overrides,
	});

	const prefsMutation = useMutation({
		mutationFn: () => updateSettings(buildPayload()),
		onSuccess: (updated) => {
			saveUserPrefs({
				dateFormat: updated.date_format as DateFormat,
				timeFormat: updated.time_format as TimeFormat,
				timezone: updated.timezone,
				numberFormat: updated.number_format as NumberFormat,
			});
			setRetentionDays(updated.audit_log_retention_days ?? 0);
			queryClient.setQueryData(["settings"], updated);
		},
	});

	const auditMutation = useMutation({
		mutationFn: () =>
			updateSettings(buildPayload({ audit_log_retention_days: retentionDays })),
		onSuccess: (updated) => {
			setRetentionDays(updated.audit_log_retention_days ?? 0);
			queryClient.setQueryData(["settings"], updated);
		},
	});

	const networkMutation = useMutation({
		mutationFn: () =>
			updateSettings(
				buildPayload({
					network_color_by: networkDefaults.colorBy,
					network_show_avatar: networkDefaults.showAvatar,
					network_show_only_mine: networkDefaults.showOnlyMine,
					network_show_unconnected: networkDefaults.showUnconnected,
					network_only_mine_depth: networkDefaults.onlyMineDepth,
				}),
			),
		onSuccess: (updated) => {
			saveUserPrefs({
				networkColorBy: updated.network_color_by,
				networkShowAvatar: updated.network_show_avatar,
				networkShowOnlyMine: updated.network_show_only_mine,
				networkShowUnconnected: updated.network_show_unconnected,
				networkOnlyMineDepth: (updated.network_only_mine_depth ??
					"direct") as NetworkOnlyMineDepth,
			});
			queryClient.setQueryData(["settings"], updated);
		},
	});

	const cleanupMutation = useMutation({
		mutationFn: runAuditCleanup,
	});

	return (
		<div className="space-y-6 max-w-md">
			<h1 className="text-[18px] font-semibold tracking-tight text-ink font-display">
				{t("general.title")}
			</h1>

			<LanguageSettingsCard
				apiSettings={apiSettings}
				isPlaceholderData={isPlaceholderData}
				buildPayload={buildPayload}
				queryClient={queryClient}
			/>

			<Card>
				<CardHeader>
					<CardTitle className="text-[14px] font-medium">
						{t("general.dateTime.title")}
					</CardTitle>
					<CardDescription className="text-[12px]">
						{t("general.dateTime.description")}
					</CardDescription>
				</CardHeader>
				<CardContent className="space-y-5">
					{/* Date format */}
					<div className="space-y-2">
						<Label className="text-[13px]">
							{t("general.dateTime.dateFormat")}
						</Label>
						<div className="space-y-1.5">
							{DATE_FORMAT_OPTIONS.map((opt) => (
								<label
									key={opt.value}
									className="flex items-center gap-3 cursor-pointer"
								>
									<input
										type="radio"
										name="dateFormat"
										value={opt.value}
										checked={prefs.dateFormat === opt.value}
										onChange={() =>
											setPrefs((p) => ({ ...p, dateFormat: opt.value }))
										}
										className="accent-accent"
									/>
									<span className="text-[13px] text-ink">
										{t(opt.labelKey)}
									</span>
									<span className="font-mono text-[11px] text-sub">
										{opt.example}
									</span>
								</label>
							))}
						</div>
					</div>

					{/* Time format */}
					<div className="space-y-2">
						<Label className="text-[13px]">
							{t("general.dateTime.timeFormat")}
						</Label>
						<div className="space-y-1.5">
							{TIME_FORMAT_OPTIONS.map((opt) => (
								<label
									key={opt.value}
									className="flex items-center gap-3 cursor-pointer"
								>
									<input
										type="radio"
										name="timeFormat"
										value={opt.value}
										checked={prefs.timeFormat === opt.value}
										onChange={() =>
											setPrefs((p) => ({ ...p, timeFormat: opt.value }))
										}
										className="accent-accent"
									/>
									<span className="text-[13px] text-ink">
										{t(opt.labelKey)}
									</span>
									<span className="font-mono text-[11px] text-sub">
										{opt.example}
									</span>
								</label>
							))}
						</div>
					</div>

					{/* Number format */}
					<div className="space-y-2">
						<Label className="text-[13px]">
							{t("general.dateTime.numberFormat")}
						</Label>
						<div className="space-y-1.5">
							{NUMBER_FORMAT_OPTIONS.map((opt) => (
								<label
									key={opt.value}
									className="flex items-center gap-3 cursor-pointer"
								>
									<input
										type="radio"
										name="numberFormat"
										value={opt.value}
										checked={prefs.numberFormat === opt.value}
										onChange={() =>
											setPrefs((p) => ({ ...p, numberFormat: opt.value }))
										}
										className="accent-accent"
									/>
									<span className="text-[13px] text-ink">
										{t(opt.labelKey)}
									</span>
									<span className="font-mono text-[11px] text-sub">
										{opt.example}
									</span>
								</label>
							))}
						</div>
					</div>

					{/* Timezone */}
					<div className="space-y-2">
						<Label className="text-[13px]">
							{t("general.dateTime.timezone")}
						</Label>
						<input
							list="tz-list"
							value={prefs.timezone}
							onChange={(e) =>
								setPrefs((p) => ({ ...p, timezone: e.target.value }))
							}
							placeholder={t("general.dateTime.timezonePlaceholder")}
							className="h-9 w-full border-field-bw border-field-line rounded-base bg-field px-3 text-[13px] focus:outline-none focus:ring-2 focus:ring-ring"
						/>
						<datalist id="tz-list">
							{COMMON_TIMEZONES.map((tz) => (
								<option key={tz} value={tz} />
							))}
						</datalist>
						<p className="text-[11px] text-sub">
							{t("general.dateTime.timezoneHint")}
						</p>
					</div>

					<Button
						onClick={() => prefsMutation.mutate()}
						size="sm"
						disabled={isPlaceholderData || prefsMutation.isPending}
					>
						{prefsMutation.isPending
							? t("common.saving")
							: prefsMutation.isSuccess
								? t("common.saved")
								: t("general.dateTime.save")}
					</Button>
					{prefsMutation.isError && (
						<p className="text-[12px] text-danger-fg">
							{t("common.saveFailed")}
						</p>
					)}
				</CardContent>
			</Card>

			<Card>
				<CardHeader>
					<CardTitle className="text-[14px] font-medium">
						{t("general.audit.title")}
					</CardTitle>
					<CardDescription className="text-[12px]">
						{t("general.audit.description")}
					</CardDescription>
				</CardHeader>
				<CardContent className="space-y-4">
					<div className="space-y-2">
						<Label className="text-[13px]">
							{t("general.audit.retention")}
						</Label>
						<input
							type="number"
							min={0}
							value={retentionDays}
							onChange={(e) =>
								setRetentionDays(Math.max(0, parseInt(e.target.value, 10) || 0))
							}
							className="h-9 w-32 border-field-bw border-field-line rounded-base bg-field px-3 text-[13px] focus:outline-none focus:ring-2 focus:ring-ring"
						/>
						<p className="text-[11px] text-sub">
							{t("general.audit.retentionHint")}
						</p>
					</div>

					<div className="flex items-center gap-3">
						<Button
							onClick={() => auditMutation.mutate()}
							size="sm"
							disabled={isPlaceholderData || auditMutation.isPending}
						>
							{auditMutation.isPending
								? t("common.saving")
								: auditMutation.isSuccess
									? t("common.saved")
									: t("common.save")}
						</Button>
						<Button
							variant="outline"
							size="sm"
							onClick={() => cleanupMutation.mutate()}
							disabled={cleanupMutation.isPending || retentionDays === 0}
						>
							{cleanupMutation.isPending
								? t("general.audit.running")
								: t("general.audit.runCleanup")}
						</Button>
					</div>

					{cleanupMutation.isSuccess && (
						<p className="text-[12px] text-sub">
							{t("general.audit.deleted", {
								count: cleanupMutation.data.deleted,
							})}
						</p>
					)}
					{cleanupMutation.isError && (
						<p className="text-[12px] text-danger-fg">
							{t("general.audit.cleanupFailed")}
						</p>
					)}
				</CardContent>
			</Card>

			<Card>
				<CardHeader>
					<CardTitle className="text-[14px] font-medium">
						{t("general.network.title")}
					</CardTitle>
					<CardDescription className="text-[12px]">
						{t("general.network.description")}
					</CardDescription>
				</CardHeader>
				<CardContent className="space-y-4">
					<div className="space-y-2">
						<Label className="text-[13px]">
							{t("general.network.colorBy")}
						</Label>
						<div className="flex gap-4">
							{(["labels", "type"] as const).map((val) => (
								<label
									key={val}
									className="flex items-center gap-2 cursor-pointer"
								>
									<input
										type="radio"
										name="networkColorBy"
										value={val}
										checked={networkDefaults.colorBy === val}
										onChange={() =>
											setNetworkDefaults((n) => ({ ...n, colorBy: val }))
										}
										className="accent-accent"
									/>
									<span className="text-[13px] text-ink">
										{val === "labels"
											? t("general.network.colorLabels")
											: t("general.network.colorType")}
									</span>
								</label>
							))}
						</div>
					</div>

					<div className="space-y-2">
						<Label className="text-[13px]">
							{t("general.network.visibility")}
						</Label>
						<div className="space-y-1.5">
							{(
								[
									{
										key: "showAvatar" as const,
										label: t("general.network.showAvatars"),
									},
									{
										key: "showOnlyMine" as const,
										label: t("general.network.showOnlyMine"),
									},
									{
										key: "showUnconnected" as const,
										label: t("general.network.showUnconnected"),
									},
								] satisfies {
									key: keyof typeof networkDefaults;
									label: string;
								}[]
							).map(({ key, label }) => (
								<div key={key}>
									<label className="flex items-center gap-3 cursor-pointer">
										<input
											type="checkbox"
											checked={networkDefaults[key] as boolean}
											onChange={(e) =>
												setNetworkDefaults((n) => ({
													...n,
													[key]: e.target.checked,
												}))
											}
											className="accent-accent"
										/>
										<span className="text-[13px] text-ink">{label}</span>
									</label>
									{key === "showOnlyMine" && networkDefaults.showOnlyMine && (
										<div className="ml-6 mt-1.5 flex gap-4">
											{(["direct", "alter"] as NetworkOnlyMineDepth[]).map(
												(v) => (
													<label
														key={v}
														className="flex items-center gap-2 cursor-pointer"
													>
														<input
															type="radio"
															name="onlyMineDepth"
															value={v}
															checked={networkDefaults.onlyMineDepth === v}
															onChange={() =>
																setNetworkDefaults((n) => ({
																	...n,
																	onlyMineDepth: v,
																}))
															}
															className="accent-accent"
														/>
														<span className="text-[12px] text-sub">
															{v === "direct"
																? "Direct only"
																: "Include alters (2nd level)"}
														</span>
													</label>
												),
											)}
										</div>
									)}
								</div>
							))}
						</div>
					</div>

					<Button
						onClick={() => networkMutation.mutate()}
						size="sm"
						disabled={isPlaceholderData || networkMutation.isPending}
					>
						{networkMutation.isPending
							? t("common.saving")
							: networkMutation.isSuccess
								? t("common.saved")
								: t("general.network.save")}
					</Button>
					{networkMutation.isError && (
						<p className="text-[12px] text-danger-fg">
							{t("common.saveFailed")}
						</p>
					)}
				</CardContent>
			</Card>

			<FavoritesSettingsCard
				apiSettings={apiSettings}
				isPlaceholderData={isPlaceholderData}
				buildPayload={buildPayload}
				queryClient={queryClient}
			/>

			<PageSizeSettingsCard
				apiSettings={apiSettings}
				isPlaceholderData={isPlaceholderData}
				buildPayload={buildPayload}
				queryClient={queryClient}
			/>

			<RelationshipDisplaySettingsCard
				apiSettings={apiSettings}
				isPlaceholderData={isPlaceholderData}
				buildPayload={buildPayload}
				queryClient={queryClient}
			/>

			<DashboardWidgetsSettingsCard
				apiSettings={apiSettings}
				isPlaceholderData={isPlaceholderData}
				buildPayload={buildPayload}
				queryClient={queryClient}
			/>

			<SearchScopeSettingsCard
				apiSettings={apiSettings}
				isPlaceholderData={isPlaceholderData}
				buildPayload={buildPayload}
				queryClient={queryClient}
			/>
		</div>
	);
}
