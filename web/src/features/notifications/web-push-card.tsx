import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
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
import {
	listNotificationChannels,
	rotateWebPushKeys,
} from "#/endpoints/notifications";
import { apiErrorMessage } from "#/lib/api-error-message";
import { keys } from "#/query-keys";
import {
	isPushSupported,
	ownedChannelId,
	subscribeThisBrowser,
	unsubscribeThisBrowser,
} from "./web-push";

export function WebPushCard() {
	const { t } = useTranslation("settings");
	const qc = useQueryClient();
	const [confirmRotate, setConfirmRotate] = useState(false);
	const channels = useQuery({
		queryKey: keys.notifications.channels,
		queryFn: listNotificationChannels,
	});
	const refresh = () =>
		qc.invalidateQueries({ queryKey: keys.notifications.channels });

	const supported = isPushSupported();
	const available = channels.data?.kinds.includes("webpush") ?? false;
	const mine = ownedChannelId(channels.data?.channels ?? []);
	const subscribed = mine !== null;

	const subscribe = useMutation({
		mutationFn: () =>
			subscribeThisBrowser(t("notifications.webpush.defaultName")),
		onSuccess: refresh,
	});
	const unsubscribe = useMutation({
		mutationFn: () => unsubscribeThisBrowser(mine),
		onSuccess: refresh,
	});
	const rotate = useMutation({
		mutationFn: rotateWebPushKeys,
		onSuccess: () => {
			setConfirmRotate(false);
			// local cleanup must not turn a successful rotation into an error
			void unsubscribeThisBrowser(null)
				.catch(() => {})
				.finally(refresh);
		},
	});

	if (!available) return null;

	const error = subscribe.error ?? unsubscribe.error ?? rotate.error;
	const denied = supported && Notification.permission === "denied";

	return (
		<Card>
			<CardHeader>
				<CardTitle className="text-[14px] font-medium">
					{t("notifications.webpush.title")}
				</CardTitle>
				<CardDescription className="text-[12px]">
					{t("notifications.webpush.description")}
				</CardDescription>
			</CardHeader>
			<CardContent className="space-y-3">
				{!supported ? (
					<p className="text-[12px] text-sub">
						{t("notifications.webpush.unsupported")}
					</p>
				) : (
					<div className="flex flex-wrap items-center gap-3">
						{subscribed ? (
							<Button
								variant="outline"
								size="sm"
								disabled={unsubscribe.isPending}
								onClick={() => unsubscribe.mutate()}
							>
								{t("notifications.webpush.disable")}
							</Button>
						) : (
							<Button
								size="sm"
								disabled={denied || subscribe.isPending}
								onClick={() => subscribe.mutate()}
							>
								{t("notifications.webpush.enable")}
							</Button>
						)}
						{confirmRotate ? (
							<>
								<Button
									variant="destructive"
									size="sm"
									disabled={rotate.isPending}
									onClick={() => rotate.mutate()}
								>
									{t("notifications.webpush.rotateConfirm")}
								</Button>
								<Button
									variant="outline"
									size="sm"
									onClick={() => setConfirmRotate(false)}
								>
									{t("common.cancel")}
								</Button>
							</>
						) : (
							<Button
								variant="outline"
								size="sm"
								onClick={() => setConfirmRotate(true)}
							>
								{t("notifications.webpush.rotate")}
							</Button>
						)}
					</div>
				)}
				{denied && (
					<p className="text-[12px] text-danger-fg">
						{t("notifications.webpush.denied")}
					</p>
				)}
				{confirmRotate && (
					<p className="text-[12px] text-sub">
						{t("notifications.webpush.rotateWarning")}
					</p>
				)}
				{error && (
					<Alert variant="destructive">
						<AlertDescription>
							{error instanceof Error && error.message === "permission_denied"
								? t("notifications.webpush.denied")
								: apiErrorMessage(error)}
						</AlertDescription>
					</Alert>
				)}
			</CardContent>
		</Card>
	);
}
