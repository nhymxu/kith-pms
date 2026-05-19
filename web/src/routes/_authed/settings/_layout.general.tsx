import { createFileRoute } from "@tanstack/react-router"
import { useState } from "react"
import { getUserPrefs, saveUserPrefs, type DateFormat, type TimeFormat } from "#/lib/format-datetime"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "#/components/ui/card"
import { Label } from "#/components/ui/label"
import { Button } from "#/components/ui/button"

export const Route = createFileRoute("/_authed/settings/_layout/general")({
	component: GeneralSettingsPage,
})

const DATE_FORMAT_OPTIONS: { value: DateFormat; label: string; example: string }[] = [
	{ value: "YYYY-MM-DD", label: "ISO (default)", example: "2026-05-19" },
	{ value: "MM/DD/YYYY", label: "US", example: "05/19/2026" },
	{ value: "DD/MM/YYYY", label: "European", example: "19/05/2026" },
]

const TIME_FORMAT_OPTIONS: { value: TimeFormat; label: string; example: string }[] = [
	{ value: "24h", label: "24-hour (default)", example: "14:30" },
	{ value: "12h", label: "12-hour", example: "2:30 PM" },
]

// Common IANA timezones for the datalist
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
]

function GeneralSettingsPage() {
	const [prefs, setPrefs] = useState(getUserPrefs)
	const [saved, setSaved] = useState(false)

	function handleSave() {
		saveUserPrefs(prefs)
		setSaved(true)
		setTimeout(() => setSaved(false), 2000)
	}

	return (
		<div className="space-y-6 max-w-md">
			<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">General</h1>

			<Card>
				<CardHeader>
					<CardTitle className="text-[14px] font-medium text-zinc-900">Date &amp; Time</CardTitle>
					<CardDescription className="text-[12px] text-zinc-500">
						Controls how dates and times are displayed throughout the app. The backend always stores UTC.
					</CardDescription>
				</CardHeader>
				<CardContent className="space-y-5">
					{/* Date format */}
					<div className="space-y-2">
						<Label className="text-[13px]">Date format</Label>
						<div className="space-y-1.5">
							{DATE_FORMAT_OPTIONS.map((opt) => (
								<label key={opt.value} className="flex items-center gap-3 cursor-pointer">
									<input
										type="radio"
										name="dateFormat"
										value={opt.value}
										checked={prefs.dateFormat === opt.value}
										onChange={() => setPrefs((p) => ({ ...p, dateFormat: opt.value }))}
										className="accent-indigo-600"
									/>
									<span className="text-[13px] text-zinc-700">{opt.label}</span>
									<span className="font-mono text-[11px] text-zinc-400">{opt.example}</span>
								</label>
							))}
						</div>
					</div>

					{/* Time format */}
					<div className="space-y-2">
						<Label className="text-[13px]">Time format</Label>
						<div className="space-y-1.5">
							{TIME_FORMAT_OPTIONS.map((opt) => (
								<label key={opt.value} className="flex items-center gap-3 cursor-pointer">
									<input
										type="radio"
										name="timeFormat"
										value={opt.value}
										checked={prefs.timeFormat === opt.value}
										onChange={() => setPrefs((p) => ({ ...p, timeFormat: opt.value }))}
										className="accent-indigo-600"
									/>
									<span className="text-[13px] text-zinc-700">{opt.label}</span>
									<span className="font-mono text-[11px] text-zinc-400">{opt.example}</span>
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
							onChange={(e) => setPrefs((p) => ({ ...p, timezone: e.target.value }))}
							placeholder="e.g. Asia/Saigon"
							className="h-9 w-full border border-zinc-200 rounded-md bg-white px-3 text-[13px] focus:outline-none focus:ring-2 focus:ring-indigo-600"
						/>
						<datalist id="tz-list">
							{COMMON_TIMEZONES.map((tz) => <option key={tz} value={tz} />)}
						</datalist>
						<p className="text-[11px] text-zinc-400">Used for display only. Default: UTC.</p>
					</div>

					<Button onClick={handleSave} size="sm">
						{saved ? "Saved!" : "Save preferences"}
					</Button>
				</CardContent>
			</Card>
		</div>
	)
}
