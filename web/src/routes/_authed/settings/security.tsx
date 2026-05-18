import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { useForm } from "@tanstack/react-form"
import { useState } from "react"
import { z } from "zod"
import { changePassword, logoutAll } from "#/endpoints/auth"
import { useAuth } from "#/lib/auth-context"
import { useQueryClient } from "@tanstack/react-query"
import { FormField } from "#/components/form/form-field"
import { SubmitButton } from "#/components/form/submit-button"
import { Button } from "#/components/ui/button"
import { Alert, AlertDescription } from "#/components/ui/alert"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "#/components/ui/card"

export const Route = createFileRoute("/_authed/settings/security")({
	component: SecurityPage,
})

const changePasswordSchema = z.object({
	current_password: z.string().min(1, "Current password is required"),
	new_password: z.string().min(8, "New password must be at least 8 characters"),
	confirm_password: z.string().min(1, "Please confirm your new password"),
}).refine((d) => d.new_password === d.confirm_password, {
	message: "Passwords do not match",
	path: ["confirm_password"],
})

type ChangePasswordValues = z.infer<typeof changePasswordSchema>

function ChangePasswordForm() {
	const navigate = useNavigate()
	const qc = useQueryClient()
	const { logout } = useAuth()
	const [apiError, setApiError] = useState<string | null>(null)
	const [isRateLimited, setIsRateLimited] = useState(false)
	const [success, setSuccess] = useState(false)

	const form = useForm({
		defaultValues: {
			current_password: "",
			new_password: "",
			confirm_password: "",
		} satisfies ChangePasswordValues,
		validators: {
			onSubmit: ({ value }) => {
				const r = changePasswordSchema.safeParse(value)
				return r.success ? undefined : r.error.issues[0]?.message
			},
		},
		onSubmit: async ({ value }) => {
			setApiError(null)
			setIsRateLimited(false)
			try {
				await changePassword(value.current_password, value.new_password, value.confirm_password)
				setSuccess(true)
				form.reset()
			} catch (err) {
				const msg = err instanceof Error ? err.message : "Failed to change password"
				if (msg.includes("429") || msg.toLowerCase().includes("too many")) {
					setIsRateLimited(true)
				} else {
					setApiError(msg)
				}
			}
		},
	})

	if (success) {
		return (
			<div className="space-y-3">
				<Alert>
					<AlertDescription className="text-[13px]">
						Password changed successfully. You may want to log out all other sessions.
					</AlertDescription>
				</Alert>
				<div className="flex gap-2">
					<Button variant="neutral" size="sm" onClick={() => setSuccess(false)}>
						Change again
					</Button>
					<LogoutAllButton onDone={() => {
						qc.clear()
						logout().then(() => navigate({ to: "/login" }))
					}} />
				</div>
			</div>
		)
	}

	return (
		<form onSubmit={(e) => { e.preventDefault(); form.handleSubmit() }} className="space-y-4">
			{isRateLimited && (
				<Alert variant="destructive">
					<AlertDescription>
						Too many attempts. Please wait before trying again.
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
						label="Current password"
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
						label="New password"
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
						label="Confirm new password"
						type="password"
						autoComplete="new-password"
						placeholder="••••••••"
					/>
				)}
			</form.Field>
			<form.Subscribe selector={(s) => s.isSubmitting}>
				{(isSubmitting) => (
					<SubmitButton isPending={isSubmitting} pendingLabel="Changing…">
						Change password
					</SubmitButton>
				)}
			</form.Subscribe>
		</form>
	)
}

function LogoutAllButton({ onDone }: { onDone: () => void }) {
	const [isPending, setIsPending] = useState(false)
	const [err, setErr] = useState<string | null>(null)

	async function handleLogoutAll() {
		setIsPending(true)
		setErr(null)
		try {
			await logoutAll()
			onDone()
		} catch (e) {
			setErr(e instanceof Error ? e.message : "Failed")
			setIsPending(false)
		}
	}

	return (
		<div className="space-y-1">
			<Button variant="destructive" size="sm" onClick={handleLogoutAll} disabled={isPending}>
				{isPending ? "Logging out…" : "Log out all sessions"}
			</Button>
			{err && <p className="text-[11px] text-red-600">{err}</p>}
		</div>
	)
}

function SecurityPage() {
	return (
		<div className="space-y-6 max-w-md">
			<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">Security</h1>

			<Card>
				<CardHeader>
					<CardTitle className="text-[14px] font-medium text-zinc-900">Change password</CardTitle>
					<CardDescription className="text-[12px] text-zinc-500">Update your login password.</CardDescription>
				</CardHeader>
				<CardContent>
					<ChangePasswordForm />
				</CardContent>
			</Card>
		</div>
	)
}
