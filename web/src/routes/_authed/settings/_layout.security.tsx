import { useForm } from "@tanstack/react-form";
import { useQueryClient } from "@tanstack/react-query";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { FormField } from "#/components/form/form-field";
import { SubmitButton } from "#/components/form/submit-button";
import { Alert, AlertDescription } from "#/components/ui/alert";
import { Button } from "#/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "#/components/ui/card";
import { changePassword, logoutAll } from "#/endpoints/auth";
import { ApiError } from "#/lib/api-client";
import { apiErrorMessage } from "#/lib/api-error-message";
import { useAuth } from "#/lib/auth-context";
import i18n from "#/lib/i18n";

export const Route = createFileRoute("/_authed/settings/_layout/security")({
	component: SecurityPage,
});

const changePasswordSchema = z
	.object({
		current_password: z.string().min(1, {
			error: () => i18n.t("errors:validation.currentPasswordRequired"),
		}),
		new_password: z
			.string()
			.min(8, { error: () => i18n.t("errors:validation.newPasswordTooShort") }),
		confirm_password: z.string().min(1, {
			error: () => i18n.t("errors:validation.confirmPasswordRequired"),
		}),
	})
	.refine((d) => d.new_password === d.confirm_password, {
		error: () => i18n.t("errors:validation.passwordsMismatch"),
		path: ["confirm_password"],
	});

type ChangePasswordValues = z.infer<typeof changePasswordSchema>;

function ChangePasswordForm() {
	const { t } = useTranslation("settings");
	const navigate = useNavigate();
	const qc = useQueryClient();
	const { logout } = useAuth();
	const [apiError, setApiError] = useState<string | null>(null);
	const [isRateLimited, setIsRateLimited] = useState(false);
	const [success, setSuccess] = useState(false);

	const form = useForm({
		defaultValues: {
			current_password: "",
			new_password: "",
			confirm_password: "",
		} satisfies ChangePasswordValues,
		validators: {
			onSubmit: ({ value }) => {
				const r = changePasswordSchema.safeParse(value);
				return r.success ? undefined : r.error.issues[0]?.message;
			},
		},
		onSubmit: async ({ value }) => {
			setApiError(null);
			setIsRateLimited(false);
			try {
				await changePassword(
					value.current_password,
					value.new_password,
					value.confirm_password,
				);
				setSuccess(true);
				form.reset();
			} catch (err) {
				if (err instanceof ApiError && err.status === 429) {
					setIsRateLimited(true);
				} else {
					setApiError(
						apiErrorMessage(err, t("security.changePassword.failed")),
					);
				}
			}
		},
	});

	if (success) {
		return (
			<div className="space-y-3">
				<Alert>
					<AlertDescription className="text-[13px]">
						{t("security.changePassword.success")}
					</AlertDescription>
				</Alert>
				<div className="flex gap-2">
					<Button variant="neutral" size="sm" onClick={() => setSuccess(false)}>
						{t("security.changePassword.changeAgain")}
					</Button>
					<LogoutAllButton
						onDone={() => {
							qc.clear();
							logout().then(() => navigate({ to: "/login" }));
						}}
					/>
				</div>
			</div>
		);
	}

	return (
		<form
			onSubmit={(e) => {
				e.preventDefault();
				form.handleSubmit();
			}}
			className="space-y-4"
		>
			{isRateLimited && (
				<Alert variant="destructive">
					<AlertDescription>
						{t("security.changePassword.rateLimited")}
					</AlertDescription>
				</Alert>
			)}
			{apiError && (
				<Alert variant="destructive">
					<AlertDescription>{apiError}</AlertDescription>
				</Alert>
			)}
			<form.Field name="current_password">
				{(f) => (
					<FormField
						field={f}
						label={t("security.changePassword.current")}
						type="password"
						autoComplete="current-password"
						placeholder="••••••••"
					/>
				)}
			</form.Field>
			<form.Field name="new_password">
				{(f) => (
					<FormField
						field={f}
						label={t("security.changePassword.new")}
						type="password"
						autoComplete="new-password"
						placeholder="••••••••"
					/>
				)}
			</form.Field>
			<form.Field name="confirm_password">
				{(f) => (
					<FormField
						field={f}
						label={t("security.changePassword.confirm")}
						type="password"
						autoComplete="new-password"
						placeholder="••••••••"
					/>
				)}
			</form.Field>
			<form.Subscribe selector={(s) => s.isSubmitting}>
				{(isSubmitting) => (
					<SubmitButton
						isPending={isSubmitting}
						pendingLabel={t("security.changePassword.pending")}
					>
						{t("security.changePassword.submit")}
					</SubmitButton>
				)}
			</form.Subscribe>
		</form>
	);
}

function LogoutAllButton({ onDone }: { onDone: () => void }) {
	const { t } = useTranslation("settings");
	const [isPending, setIsPending] = useState(false);
	const [err, setErr] = useState<string | null>(null);

	async function handleLogoutAll() {
		setIsPending(true);
		setErr(null);
		try {
			await logoutAll();
			onDone();
		} catch (e) {
			setErr(apiErrorMessage(e, t("security.logoutFailed")));
			setIsPending(false);
		}
	}

	return (
		<div className="space-y-1">
			<Button
				variant="destructive"
				size="sm"
				onClick={handleLogoutAll}
				disabled={isPending}
			>
				{isPending ? t("security.loggingOut") : t("security.logoutAll")}
			</Button>
			{err && <p className="text-[11px] text-danger-fg">{err}</p>}
		</div>
	);
}

function SecurityPage() {
	const { t } = useTranslation("settings");
	return (
		<div className="space-y-6 max-w-md">
			<h1 className="text-[18px] font-semibold tracking-tight text-ink font-display">
				{t("security.title")}
			</h1>

			<Card>
				<CardHeader>
					<CardTitle className="text-[14px] font-medium">
						{t("security.changePassword.title")}
					</CardTitle>
					<CardDescription className="text-[12px]">
						{t("security.changePassword.description")}
					</CardDescription>
				</CardHeader>
				<CardContent>
					<ChangePasswordForm />
				</CardContent>
			</Card>
		</div>
	);
}
