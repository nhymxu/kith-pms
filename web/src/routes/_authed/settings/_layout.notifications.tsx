import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Alert, AlertDescription } from "#/components/ui/alert";
import { Badge } from "#/components/ui/badge";
import { Button } from "#/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "#/components/ui/card";
import { Input } from "#/components/ui/input";
import { Label } from "#/components/ui/label";
import { Switch } from "#/components/ui/switch";
import {
	deleteNotificationChannel,
	listNotificationChannels,
	setNotificationChannelEnabled,
	testNotificationChannel,
} from "#/endpoints/notifications";
import { getSettings, updateSettings } from "#/endpoints/settings";
import { apiErrorMessage } from "#/lib/api-error-message";
import { saveUserPrefs } from "#/lib/format-datetime";
import { keys } from "#/query-keys";
import type {
	ChannelTestResult,
	NotificationChannel,
} from "#/schemas/notification-channel";

export const Route = createFileRoute("/_authed/settings/_layout/notifications")(
	{ component: NotificationsPage },
);

function NotifyHourCard() {
	const { t } = useTranslation("settings");
	const qc = useQueryClient();
	const { data: settings } = useQuery({
		queryKey: ["settings"],
		queryFn: getSettings,
	});
	const [draft, setDraft] = useState<string | null>(null);
	const value = draft ?? String(settings?.notify_hour ?? 9);
	const hour = Number(value);
	const valid =
		value !== "" && Number.isInteger(hour) && hour >= 0 && hour <= 23;

	const save = useMutation({
		mutationFn: () => {
			if (!settings) throw new Error("settings not loaded");
			return updateSettings({ ...settings, notify_hour: hour });
		},
		onSuccess: (updated) => {
			saveUserPrefs({ notifyHour: updated.notify_hour });
			qc.setQueryData(["settings"], updated);
			setDraft(null);
		},
	});

	return (
		<Card>
			<CardHeader>
				<CardTitle className="text-[14px] font-medium">
					{t("notifications.hour.title")}
				</CardTitle>
				<CardDescription className="text-[12px]">
					{t("notifications.hour.description", {
						timezone: settings?.timezone ?? "UTC",
					})}
				</CardDescription>
			</CardHeader>
			<CardContent className="space-y-3">
				<div className="flex items-end gap-3">
					<div className="space-y-1.5">
						<Label htmlFor="notify-hour">{t("notifications.hour.label")}</Label>
						<Input
							id="notify-hour"
							type="number"
							min={0}
							max={23}
							step={1}
							className="w-24"
							value={value}
							onChange={(e) => setDraft(e.target.value)}
						/>
					</div>
					<Button
						size="sm"
						disabled={!settings || !valid || draft === null || save.isPending}
						onClick={() => save.mutate()}
					>
						{save.isPending ? t("common.saving") : t("common.save")}
					</Button>
				</div>
				{!valid && (
					<p className="text-[11px] text-danger-fg">
						{t("notifications.hour.invalid")}
					</p>
				)}
				{save.isError && (
					<p className="text-[11px] text-danger-fg">
						{apiErrorMessage(save.error, t("common.saveFailed"))}
					</p>
				)}
			</CardContent>
		</Card>
	);
}

function ChannelRow({ channel }: { channel: NotificationChannel }) {
	const { t } = useTranslation("settings");
	const qc = useQueryClient();
	const [confirming, setConfirming] = useState(false);
	const [result, setResult] = useState<ChannelTestResult | null>(null);
	const refresh = () =>
		qc.invalidateQueries({ queryKey: keys.notifications.channels });

	// Only the latest action's outcome is shown.
	const clearFeedback = () => {
		setResult(null);
		toggle.reset();
		remove.reset();
		test.reset();
	};

	const toggle = useMutation({
		mutationFn: (enabled: boolean) =>
			setNotificationChannelEnabled(channel.id, enabled),
		onMutate: clearFeedback,
		onSuccess: refresh,
	});
	const remove = useMutation({
		mutationFn: () => deleteNotificationChannel(channel.id),
		onMutate: clearFeedback,
		onSuccess: refresh,
	});
	const test = useMutation({
		mutationFn: () => testNotificationChannel(channel.id),
		onMutate: clearFeedback,
		onSuccess: setResult,
	});

	const unreadable = channel.status === "secret_unreadable";
	const failure = result && !result.ok;
	const error = toggle.error ?? remove.error ?? test.error;

	return (
		<li className="space-y-2 py-3">
			<div className="flex flex-wrap items-center gap-3">
				<div className="min-w-0 flex-1">
					<p className="truncate text-[13px] font-medium text-ink">
						{channel.name || channel.kind}
					</p>
					<div className="mt-1 flex gap-1.5">
						<Badge variant="neutral">{channel.kind}</Badge>
						{unreadable && (
							<Badge variant="destructive">
								{t("notifications.channels.unreadable")}
							</Badge>
						)}
					</div>
				</div>
				<Switch
					aria-label={t("notifications.channels.enabled")}
					checked={channel.enabled}
					disabled={toggle.isPending}
					onCheckedChange={(v) => toggle.mutate(v === true)}
				/>
				<Button
					variant="outline"
					size="sm"
					disabled={unreadable || test.isPending}
					onClick={() => test.mutate()}
				>
					{test.isPending
						? t("notifications.channels.testing")
						: t("notifications.channels.test")}
				</Button>
				{confirming ? (
					<>
						<Button
							variant="destructive"
							size="sm"
							disabled={remove.isPending}
							onClick={() => remove.mutate()}
						>
							{t("notifications.channels.deleteConfirm")}
						</Button>
						<Button
							variant="outline"
							size="sm"
							onClick={() => setConfirming(false)}
						>
							{t("common.cancel")}
						</Button>
					</>
				) : (
					<Button
						variant="outline"
						size="sm"
						onClick={() => setConfirming(true)}
					>
						{t("common.delete")}
					</Button>
				)}
			</div>
			{unreadable && (
				<p className="text-[12px] text-danger-fg">
					{t("notifications.channels.unreadableHint")}
				</p>
			)}
			{result?.ok && (
				<p className="text-[12px] text-success-fg">
					{t("notifications.channels.testOk")}
				</p>
			)}
			{failure && (
				<p className="text-[12px] text-danger-fg">
					{t(
						result.status
							? "notifications.channels.testFailedStatus"
							: "notifications.channels.testFailed",
						{
							reason: t(
								`notifications.testClass.${result.class ?? "internal"}`,
								{ defaultValue: result.class ?? "" },
							),
							status: result.status,
						},
					)}
				</p>
			)}
			{error && (
				<p className="text-[12px] text-danger-fg">{apiErrorMessage(error)}</p>
			)}
		</li>
	);
}

function NotificationsPage() {
	const { t } = useTranslation("settings");
	const channels = useQuery({
		queryKey: keys.notifications.channels,
		queryFn: listNotificationChannels,
	});

	return (
		<div className="space-y-6 max-w-2xl">
			<h1 className="text-[18px] font-semibold tracking-tight text-ink font-display">
				{t("notifications.title")}
			</h1>

			<NotifyHourCard />

			<Card>
				<CardHeader>
					<CardTitle className="text-[14px] font-medium">
						{t("notifications.channels.title")}
					</CardTitle>
					<CardDescription className="text-[12px]">
						{t("notifications.channels.description")}
					</CardDescription>
				</CardHeader>
				<CardContent>
					{channels.isError && (
						<Alert variant="destructive">
							<AlertDescription>
								{apiErrorMessage(
									channels.error,
									t("notifications.channels.loadFailed"),
								)}
							</AlertDescription>
						</Alert>
					)}
					{channels.data && channels.data.channels.length === 0 && (
						<p className="text-[13px] text-sub">
							{channels.data.kinds.length === 0
								? t("notifications.channels.noKinds")
								: t("notifications.channels.empty")}
						</p>
					)}
					{channels.data && channels.data.channels.length > 0 && (
						<ul className="divide-y divide-line">
							{channels.data.channels.map((c) => (
								<ChannelRow key={c.id} channel={c} />
							))}
						</ul>
					)}
				</CardContent>
			</Card>
		</div>
	);
}
