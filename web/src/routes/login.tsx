import { createFileRoute, useNavigate, useSearch, redirect } from "@tanstack/react-router"
import { useForm } from "@tanstack/react-form"
import { useState } from "react"
import { z } from "zod"
import { useAuth } from "#/lib/auth-context"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "#/components/ui/card"
import { FormField } from "#/components/form/form-field"
import { SubmitButton } from "#/components/form/submit-button"
import { Alert, AlertDescription } from "#/components/ui/alert"

const loginSearchSchema = z.object({
	redirect: z.string().optional(),
})

export const Route = createFileRoute("/login")({
	validateSearch: loginSearchSchema,
	beforeLoad({ context }) {
		if (!context.auth.isLoading && context.auth.user) {
			throw redirect({ to: "/" })
		}
	},
	component: LoginPage,
})

function LoginPage() {
	const { login } = useAuth()
	const navigate = useNavigate()
	const search = useSearch({ from: "/login" })
	const [apiError, setApiError] = useState<string | null>(null)

	const form = useForm({
		defaultValues: { password: "" },
		onSubmit: async ({ value }) => {
			setApiError(null)
			try {
				await login(value.password)
				const target = search.redirect ?? "/"
				// Sanitize — only allow same-origin redirects.
				const safeTarget = target.startsWith("/") ? target : "/"
				navigate({ to: safeTarget })
			} catch (err) {
				setApiError(err instanceof Error ? err.message : "Login failed")
			}
		},
	})

	return (
		<div className="min-h-screen flex items-center justify-center bg-background p-4">
			<Card className="w-full max-w-sm">
				<CardHeader>
					<CardTitle className="text-xl">Kith PMS</CardTitle>
					<CardDescription>Enter your password to continue</CardDescription>
				</CardHeader>
				<CardContent>
					<form
						onSubmit={(e) => {
							e.preventDefault()
							form.handleSubmit()
						}}
						className="space-y-4"
					>
						{apiError && (
							<Alert variant="destructive">
								<AlertDescription>{apiError}</AlertDescription>
							</Alert>
						)}

						<form.Field
							name="password"
							validators={{ onChange: ({ value }) => (!value ? "Password is required" : undefined) }}
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
								<SubmitButton isPending={isSubmitting} pendingLabel="Signing in…" className="w-full">
									Sign in
								</SubmitButton>
							)}
						</form.Subscribe>
					</form>
				</CardContent>
			</Card>
		</div>
	)
}
