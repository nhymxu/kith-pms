import { useState, type ReactNode } from "react"
import { RequireAuth } from "#/lib/auth-context"
import { Sheet, SheetContent, SheetTitle } from "#/components/ui/sheet"
import { Sidebar } from "./sidebar"
import { Topbar } from "./topbar"

interface AppShellProps {
	children: ReactNode
}

export function AppShell({ children }: AppShellProps) {
	return (
		<RequireAuth>
			<AppShellInner>{children}</AppShellInner>
		</RequireAuth>
	)
}

function AppShellInner({ children }: AppShellProps) {
	const [mobileOpen, setMobileOpen] = useState(false)

	return (
		<div className="flex h-screen overflow-hidden bg-background">
			{/* Desktop sidebar */}
			<aside className="hidden md:flex w-56 shrink-0 flex-col border-r-2 border-border bg-background">
				<Sidebar />
			</aside>

			{/* Mobile drawer */}
			<Sheet open={mobileOpen} onOpenChange={setMobileOpen}>
				<SheetContent side="left" className="w-56 p-0">
					{/* Visually hidden title satisfies Radix a11y requirement for dialog content */}
					<SheetTitle className="sr-only">Navigation</SheetTitle>
					<Sidebar onNavClick={() => setMobileOpen(false)} />
				</SheetContent>
			</Sheet>

			{/* Main content area */}
			<div className="flex flex-1 flex-col overflow-hidden">
				<Topbar onMenuClick={() => setMobileOpen(true)} />
				<main className="flex-1 overflow-y-auto p-6">
					{children}
				</main>
			</div>
		</div>
	)
}
