import { useForm } from "@tanstack/react-form";
import {
	createFileRoute,
	redirect,
	useNavigate,
	useSearch,
} from "@tanstack/react-router";
import { useState } from "react";
import { z } from "zod";
import { FormField } from "#/components/form/form-field";
import { SubmitButton } from "#/components/form/submit-button";
import { Alert, AlertDescription } from "#/components/ui/alert";
import { useAuth } from "#/lib/auth-context";

const loginSearchSchema = z.object({
	redirect: z.string().optional(),
});

export const Route = createFileRoute("/login")({
	validateSearch: loginSearchSchema,
	beforeLoad({ context }) {
		if (!context.auth.isLoading && context.auth.user) {
			throw redirect({ to: "/" });
		}
	},
	component: LoginPage,
});

function LoginPage() {
	const { login } = useAuth();
	const navigate = useNavigate();
	const search = useSearch({ from: "/login" });
	const [apiError, setApiError] = useState<string | null>(null);

	const form = useForm({
		defaultValues: { password: "" },
		onSubmit: async ({ value }) => {
			setApiError(null);
			try {
				await login(value.password);
				const target = search.redirect ?? "/";
				// Sanitize — only allow same-origin redirects.
				const safeTarget = target.startsWith("/") ? target : "/";
				navigate({ to: safeTarget });
			} catch (err) {
				setApiError(err instanceof Error ? err.message : "Login failed");
			}
		},
	});

	return (
		<div className="min-h-screen grid place-items-center bg-zinc-50 p-4">
			<div className="w-full max-w-[360px]">
				<h1 className="text-[28px] font-semibold text-center mb-8 tracking-tight">
					Kith
				</h1>
				<form
					onSubmit={(e) => {
						e.preventDefault();
						form.handleSubmit();
					}}
					className="border border-zinc-200 rounded-md bg-white p-6 space-y-4"
				>
					<p className="text-[15px] font-semibold text-zinc-900">Sign in</p>

					{apiError && (
						<Alert variant="destructive">
							<AlertDescription>{apiError}</AlertDescription>
						</Alert>
					)}

					<form.Field
						name="password"
						validators={{
							onChange: ({ value }) =>
								!value ? "Password is required" : undefined,
						}}
					>
						{(field) => (
							<FormField
								field={field}
								label="Password"
								type="password"
								placeholder="••••••••"
								autoComplete="current-password"
								autoFocus
							/>
						)}
					</form.Field>

					<form.Subscribe selector={(s) => s.isSubmitting}>
						{(isSubmitting) => (
							<SubmitButton
								isPending={isSubmitting}
								pendingLabel="Signing in…"
								className="w-full"
							>
								Sign in
							</SubmitButton>
						)}
					</form.Subscribe>
				</form>
			</div>
		</div>
	);
}
