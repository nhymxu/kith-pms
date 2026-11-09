import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { SubmitButton } from "#/components/form/submit-button";
import { Alert, AlertDescription } from "#/components/ui/alert";
import { Button } from "#/components/ui/button";
import { Checkbox } from "#/components/ui/checkbox";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import { Input } from "#/components/ui/input";
import { Label } from "#/components/ui/label";
import {
	createNotificationChannel,
	updateNotificationChannelConfig,
} from "#/endpoints/notifications";
import { apiErrorMessage } from "#/lib/api-error-message";
import { keys } from "#/query-keys";
import type { NotificationChannel } from "#/schemas/notification-channel";
import {
	CHANNEL_KINDS,
	type ChannelKind,
	generateWebhookSecret,
	isChannelKind,
} from "./channel-kinds";

type Props = {
	open: boolean;
	onClose: () => void;
	// Kinds the server has registered; unknown ones are not offered.
	kinds: string[];
	// Present when editing.
	channel?: NotificationChannel;
};

function initialValues(kind: ChannelKind | undefined): Record<string, string> {
	const out: Record<string, string> = {};
	for (const f of kind ? CHANNEL_KINDS[kind].fields : []) {
		if (f.generate) out[f.key] = generateWebhookSecret();
	}
	return out;
}

export function ChannelFormDialog({ open, onClose, kinds, channel }: Props) {
	const { t } = useTranslation("settings");
	const qc = useQueryClient();
	const available = kinds.filter(isChannelKind);
	const editing = channel !== undefined;
	const [kind, setKind] = useState(
		channel && isChannelKind(channel.kind) ? channel.kind : available[0],
	);
	const [name, setName] = useState(channel?.name ?? "");
	const [values, setValues] = useState<Record<string, string>>(() =>
		channel
			? Object.fromEntries(
					Object.entries(channel.config ?? {}).map(([k, v]) => [k, String(v)]),
				)
			: initialValues(kind),
	);
	const [cleared, setCleared] = useState<Record<string, boolean>>({});
	const spec = kind ? CHANNEL_KINDS[kind] : undefined;
	// An unreadable channel has nothing to keep, so every secret is re-entered.
	const reenter = channel?.status === "secret_unreadable";

	const mutation = useMutation({
		mutationFn: () => {
			const config: Record<string, string> = {};
			for (const f of spec?.fields ?? []) {
				const v = (values[f.key] ?? "").trim();
				if (v !== "") config[f.key] = v;
				else if (cleared[f.key]) config[f.key] = "";
			}
			return channel
				? updateNotificationChannelConfig(channel.id, config)
				: createNotificationChannel({ kind, name: name.trim(), config });
		},
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.notifications.channels });
			onClose();
		},
	});

	const missing = (spec?.fields ?? []).some((f) => {
		if (f.optional) return false;
		if (f.secret && editing && !reenter) return false;
		return (values[f.key] ?? "").trim() === "";
	});
	const canSubmit = spec !== undefined && !missing && (editing || name.trim());

	return (
		<Dialog open={open} onOpenChange={(v) => !v && onClose()}>
			<DialogContent>
				<DialogHeader>
					<DialogTitle>
						{editing
							? t("notifications.form.editTitle")
							: t("notifications.form.addTitle")}
					</DialogTitle>
					<DialogDescription>
						{kind && t(`notifications.kinds.${kind}.description`)}
					</DialogDescription>
				</DialogHeader>

				{mutation.isError && (
					<Alert variant="destructive">
						<AlertDescription>
							{apiErrorMessage(mutation.error, t("common.saveFailed"))}
						</AlertDescription>
					</Alert>
				)}

				{!editing && (
					<div className="flex flex-wrap gap-2">
						{available.map((k) => (
							<Button
								key={k}
								type="button"
								size="sm"
								variant={k === kind ? "default" : "outline"}
								onClick={() => {
									setKind(k);
									setValues(initialValues(k));
								}}
							>
								{t(`notifications.kinds.${k}.label`)}
							</Button>
						))}
					</div>
				)}

				{!editing && (
					<div className="space-y-1.5">
						<Label htmlFor="channel-name">{t("notifications.form.name")}</Label>
						<Input
							id="channel-name"
							maxLength={100}
							value={name}
							onChange={(e) => setName(e.target.value)}
						/>
					</div>
				)}

				{spec?.fields.map((f) => (
					<div key={f.key} className="space-y-1.5">
						<Label htmlFor={`channel-${f.key}`}>
							{t(`notifications.fields.${f.key}.label`)}
							{f.optional && (
								<span className="ml-1 text-sub">
									{t("notifications.form.optional")}
								</span>
							)}
						</Label>
						<Input
							id={`channel-${f.key}`}
							type={
								f.secret && !f.generate
									? "password"
									: f.type === "url"
										? "url"
										: "text"
							}
							autoComplete={f.secret ? "new-password" : "off"}
							data-1p-ignore
							placeholder={
								f.secret && editing && !reenter
									? t("notifications.form.secretKept")
									: f.placeholder
							}
							value={values[f.key] ?? ""}
							onChange={(e) =>
								setValues((prev) => ({ ...prev, [f.key]: e.target.value }))
							}
						/>
						<p className="text-[11px] text-sub">
							{t(`notifications.fields.${f.key}.hint`)}
						</p>
						{f.generate && (
							<div className="flex gap-2">
								<Button
									type="button"
									size="sm"
									variant="outline"
									disabled={(values[f.key] ?? "") === ""}
									onClick={() =>
										navigator.clipboard.writeText(values[f.key] ?? "")
									}
								>
									{t("notifications.form.copySecret")}
								</Button>
								<Button
									type="button"
									size="sm"
									variant="outline"
									onClick={() =>
										setValues((prev) => ({
											...prev,
											[f.key]: generateWebhookSecret(),
										}))
									}
								>
									{t(
										editing
											? "notifications.form.rotateSecret"
											: "notifications.form.regenerateSecret",
									)}
								</Button>
							</div>
						)}
						{f.secret && f.optional && editing && !reenter && (
							<Label className="flex items-center gap-2 text-[12px] font-normal">
								<Checkbox
									checked={cleared[f.key] === true}
									onCheckedChange={(v) =>
										setCleared((prev) => ({ ...prev, [f.key]: v === true }))
									}
								/>
								{t("notifications.form.removeSaved")}
							</Label>
						)}
					</div>
				))}

				<DialogFooter>
					<Button variant="neutral" type="button" onClick={onClose}>
						{t("common.cancel")}
					</Button>
					<SubmitButton
						type="button"
						isPending={mutation.isPending}
						disabled={!canSubmit}
						onClick={() => mutation.mutate()}
					>
						{t("common.save")}
					</SubmitButton>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
