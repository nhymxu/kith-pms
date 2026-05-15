import { createFileRoute, Outlet, redirect, useNavigate } from "@tanstack/react-router"
import { useEffect } from "react"
import { AppShell } from "#/components/app-shell/app-shell"
import { onSessionLost } from "#/lib/api-client"
import { Card, CardContent, CardHeader, CardTitle } from "#/components/ui/card"
import { Button } from "#/components/ui/button"

export const Route = createFileRoute("/_authed")({
	beforeLoad({ context, location }) {
		// Only redirect if auth has resolved and there is no user.
		// During loading we let the component handle it (RequireAuth inside AppShell does it).
		if (context.auth.isLoading) return
		if (!context.auth.user) {
			throw redirect({
				to: "/login",
				search: { redirect: location.href },
			})
		}
	},
	errorComponent: AuthedErrorBoundary,
	notFoundComponent: AuthedNotFound,
	component: AuthedLayout,
})

function AuthedLayout() {
	const navigate = useNavigate()

	// Subscribe to 401 session-lost events (fires when apiFetch gets a 401 after initial load).
	useEffect(() => {
		const unsub = onSessionLost(() => {
			navigate({ to: "/login", search: { redirect: window.location.href } })
		})
		return unsub
	}, [navigate])

	return (
		<AppShell>
			<Outlet />
		</AppShell>
	)
}

function AuthedErrorBoundary({ error }: { error: unknown }) {
	const message = error instanceof Error ? error.message : "An unexpected error occurred."
	return (
		<div className="flex items-center justify-center min-h-screen p-4 bg-background">
			<Card className="max-w-md w-full">
				<CardHeader>
					<CardTitle className="text-destructive">Something went wrong</CardTitle>
				</CardHeader>
				<CardContent className="space-y-4">
					<p className="text-sm font-base">{message}</p>
					<Button variant="neutral" onClick={() => window.location.assign("/")}>
						Go home
					</Button>
				</CardContent>
			</Card>
		</div>
	)
}

function AuthedNotFound() {
	return (
		<div className="flex flex-col items-center justify-center py-24 gap-4">
			<h1 className="text-4xl font-heading">404</h1>
			<p className="text-sm font-base text-foreground/60">Page not found.</p>
			<Button variant="neutral" asChild>
				<a href="/">Go home</a>
			</Button>
		</div>
	)
}
