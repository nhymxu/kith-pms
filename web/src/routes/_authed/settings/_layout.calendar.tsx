import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Alert, AlertDescription } from "#/components/ui/alert";
import { Button } from "#/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "#/components/ui/card";
import { Checkbox } from "#/components/ui/checkbox";
import { Input } from "#/components/ui/input";
import { Label } from "#/components/ui/label";
import {
	getCalendarFeedToken,
	regenerateCalendarFeedToken,
} from "#/endpoints/calendar";
import { apiErrorMessage } from "#/lib/api-error-message";
import { keys } from "#/query-keys";

export const Route = createFileRoute("/_authed/settings/_layout/calendar")({
	component: CalendarPage,
});

const PARTS = ["birthdays", "dates", "reminders"] as const;
type Part = (typeof PARTS)[number];

const FEEDS = [
	{ file: "birthdays.ics", labelKey: "calendar.birthdays" },
	{ file: "reminders.ics", labelKey: "calendar.reminders" },
] as const;

function FeedRow({
	token,
	file,
	label,
	query = "",
}: {
	token: string;
	file: string;
	label: string;
	query?: string;
}) {
	const { t } = useTranslation("settings");
	const [copied, setCopied] = useState(false);
	const path = `/calendar/${token}/${file}${query}`;
	const httpUrl = `${window.location.origin}${path}`;
	const webcalUrl = `webcal://${window.location.host}${path}`;

	async function copy() {
		// navigator.clipboard is undefined on plain-http installs; the input stays selectable.
		if (!navigator.clipboard) return;
		await navigator.clipboard.writeText(httpUrl);
		setCopied(true);
		setTimeout(() => setCopied(false), 2000);
	}

	return (
		<div className="space-y-1.5">
			<p className="text-[13px] font-medium text-ink">{label}</p>
			<div className="flex gap-2">
				<Input readOnly value={httpUrl} className="font-mono text-[12px]" />
				<Button variant="outline" size="sm" onClick={copy}>
					{copied ? t("calendar.copied") : t("calendar.copy")}
				</Button>
				<Button variant="outline" size="sm" asChild>
					<a href={webcalUrl}>{t("calendar.subscribe")}</a>
				</Button>
			</div>
		</div>
	);
}

function CalendarPage() {
	const { t } = useTranslation("settings");
	const qc = useQueryClient();
	const [confirming, setConfirming] = useState(false);
	const [parts, setParts] = useState<Part[]>([...PARTS]);

	function togglePart(part: Part, on: boolean) {
		setParts((cur) => {
			const next = on ? [...cur, part] : cur.filter((p) => p !== part);
			return next.length === 0 ? cur : PARTS.filter((p) => next.includes(p));
		});
	}

	const combinedQuery =
		parts.length === PARTS.length ? "" : `?include=${parts.join(",")}`;

	const tokenQuery = useQuery({
		queryKey: keys.calendar.feedToken,
		queryFn: getCalendarFeedToken,
	});

	const regenerate = useMutation({
		mutationFn: regenerateCalendarFeedToken,
		onSuccess: (token) => {
			qc.setQueryData(keys.calendar.feedToken, token);
			setConfirming(false);
		},
	});

	return (
		<div className="space-y-6 max-w-2xl">
			<h1 className="text-[18px] font-semibold tracking-tight text-ink font-display">
				{t("calendar.title")}
			</h1>

			<Card>
				<CardHeader>
					<CardTitle className="text-[14px] font-medium">
						{t("calendar.title")}
					</CardTitle>
					<CardDescription className="text-[12px]">
						{t("calendar.description")}
					</CardDescription>
				</CardHeader>
				<CardContent className="space-y-4">
					{tokenQuery.isError && (
						<Alert variant="destructive">
							<AlertDescription>
								{apiErrorMessage(tokenQuery.error, t("calendar.loadFailed"))}
							</AlertDescription>
						</Alert>
					)}

					{tokenQuery.data && (
						<div className="space-y-3">
							<div className="flex flex-wrap gap-4">
								{PARTS.map((part) => (
									<div key={part} className="flex items-center gap-2">
										<Checkbox
											id={`cal-part-${part}`}
											checked={parts.includes(part)}
											onCheckedChange={(v) => togglePart(part, v === true)}
										/>
										<Label htmlFor={`cal-part-${part}`}>
											{t(`calendar.parts.${part}`)}
										</Label>
									</div>
								))}
							</div>
							<FeedRow
								token={tokenQuery.data}
								file="feed.ics"
								query={combinedQuery}
								label={t("calendar.combined")}
							/>
						</div>
					)}

					{tokenQuery.data &&
						FEEDS.map((f) => (
							<FeedRow
								key={f.file}
								token={tokenQuery.data}
								file={f.file}
								label={t(f.labelKey)}
							/>
						))}

					{tokenQuery.data && (
						<div className="space-y-2 border-t border-line pt-4">
							{confirming ? (
								<>
									<p className="text-[12px] text-sub">
										{t("calendar.regenerateWarning")}
									</p>
									<div className="flex gap-2">
										<Button
											variant="destructive"
											size="sm"
											disabled={regenerate.isPending}
											onClick={() => regenerate.mutate()}
										>
											{regenerate.isPending
												? t("calendar.regenerating")
												: t("calendar.regenerateConfirm")}
										</Button>
										<Button
											variant="outline"
											size="sm"
											disabled={regenerate.isPending}
											onClick={() => setConfirming(false)}
										>
											{t("common.cancel")}
										</Button>
									</div>
								</>
							) : (
								<Button
									variant="outline"
									size="sm"
									onClick={() => setConfirming(true)}
								>
									{t("calendar.regenerate")}
								</Button>
							)}
							{regenerate.isError && (
								<p className="text-[11px] text-danger-fg">
									{apiErrorMessage(
										regenerate.error,
										t("calendar.regenerateFailed"),
									)}
								</p>
							)}
						</div>
					)}
				</CardContent>
			</Card>
		</div>
	);
}
