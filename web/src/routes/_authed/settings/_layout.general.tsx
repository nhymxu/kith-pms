import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
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
import { getSettings, updateSettings } from "#/endpoints/settings";
import {
	type DateFormat,
	getUserPrefs,
	saveUserPrefs,
	type TimeFormat,
} from "#/lib/format-datetime";

export const Route = createFileRoute("/_authed/settings/_layout/general")({
	component: GeneralSettingsPage,
});

const DATE_FORMAT_OPTIONS: {
	value: DateFormat;
	label: string;
	example: string;
}[] = [
	{ value: "YYYY-MM-DD", label: "ISO (default)", example: "2026-05-19" },
	{ value: "MM/DD/YYYY", label: "US", example: "05/19/2026" },
	{ value: "DD/MM/YYYY", label: "European", example: "19/05/2026" },
];

const TIME_FORMAT_OPTIONS: {
	value: TimeFormat;
	label: string;
	example: string;
}[] = [
	{ value: "24h", label: "24-hour (default)", example: "14:30" },
	{ value: "12h", label: "12-hour", example: "2:30 PM" },
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
	const queryClient = useQueryClient();

	const { data: apiSettings } = useQuery({
		queryKey: ["settings"],
		queryFn: getSettings,
		// Seed local state from localStorage while the query loads.
		placeholderData: () => {
			const p = getUserPrefs();
			return {
				date_format: p.dateFormat,
				time_format: p.timeFormat,
				timezone: p.timezone,
			} as const;
		},
	});

	const [prefs, setPrefs] = useState<{
		dateFormat: DateFormat;
		timeFormat: TimeFormat;
		timezone: string;
	}>(() => {
		if (apiSettings) {
			return {
				dateFormat: apiSettings.date_format as DateFormat,
				timeFormat: apiSettings.time_format as TimeFormat,
				timezone: apiSettings.timezone,
			};
		}
		return getUserPrefs();
	});

	// Keep local form state in sync when the query resolves.
	const [synced, setSynced] = useState(false);
	if (apiSettings && !synced) {
		setPrefs({
			dateFormat: apiSettings.date_format as DateFormat,
			timeFormat: apiSettings.time_format as TimeFormat,
			timezone: apiSettings.timezone,
		});
		setSynced(true);
	}

	const mutation = useMutation({
		mutationFn: () =>
			updateSettings({
				date_format: prefs.dateFormat,
				time_format: prefs.timeFormat,
				timezone: prefs.timezone,
			}),
		onSuccess: (updated) => {
			saveUserPrefs({
				dateFormat: updated.date_format as DateFormat,
				timeFormat: updated.time_format as TimeFormat,
				timezone: updated.timezone,
			});
			queryClient.setQueryData(["settings"], updated);
		},
	});

	return (
		<div className="space-y-6 max-w-md">
			<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">
				General
			</h1>

			<Card>
				<CardHeader>
					<CardTitle className="text-[14px] font-medium text-zinc-900">
						Date &amp; Time
					</CardTitle>
					<CardDescription className="text-[12px] text-zinc-500">
						Controls how dates and times are displayed throughout the app. The
						backend always stores UTC.
					</CardDescription>
				</CardHeader>
				<CardContent className="space-y-5">
					{/* Date format */}
					<div className="space-y-2">
						<Label className="text-[13px]">Date format</Label>
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
										className="accent-indigo-600"
									/>
									<span className="text-[13px] text-zinc-700">{opt.label}</span>
									<span className="font-mono text-[11px] text-zinc-400">
										{opt.example}
									</span>
								</label>
							))}
						</div>
					</div>

					{/* Time format */}
					<div className="space-y-2">
						<Label className="text-[13px]">Time format</Label>
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
										className="accent-indigo-600"
									/>
									<span className="text-[13px] text-zinc-700">{opt.label}</span>
									<span className="font-mono text-[11px] text-zinc-400">
										{opt.example}
									</span>
								</label>
							))}
						</div>
					</div>

					{/* Timezone */}
					<div className="space-y-2">
						<Label className="text-[13px]">Timezone</Label>
						<input
							list="tz-list"
							value={prefs.timezone}
							onChange={(e) =>
								setPrefs((p) => ({ ...p, timezone: e.target.value }))
							}
							placeholder="e.g. Asia/Saigon"
							className="h-9 w-full border border-zinc-200 rounded-md bg-white px-3 text-[13px] focus:outline-none focus:ring-2 focus:ring-indigo-600"
						/>
						<datalist id="tz-list">
							{COMMON_TIMEZONES.map((tz) => (
								<option key={tz} value={tz} />
							))}
						</datalist>
						<p className="text-[11px] text-zinc-400">
							Used for display only. Default: UTC.
						</p>
					</div>

					<Button
						onClick={() => mutation.mutate()}
						size="sm"
						disabled={mutation.isPending}
					>
						{mutation.isPending
							? "Saving…"
							: mutation.isSuccess
								? "Saved!"
								: "Save preferences"}
					</Button>
					{mutation.isError && (
						<p className="text-[12px] text-red-500">
							Failed to save. Please try again.
						</p>
					)}
				</CardContent>
			</Card>
		</div>
	);
}
