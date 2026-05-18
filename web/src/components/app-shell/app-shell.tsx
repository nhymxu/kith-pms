import { type ReactNode, useState } from "react";
import { Sheet, SheetContent, SheetTitle } from "#/components/ui/sheet";
import { RequireAuth } from "#/lib/auth-context";
import { Sidebar } from "./sidebar";
import { Topbar } from "./topbar";

interface AppShellProps {
	children: ReactNode;
}

export function AppShell({ children }: AppShellProps) {
	return (
		<RequireAuth>
			<AppShellInner>{children}</AppShellInner>
		</RequireAuth>
	);
}

function AppShellInner({ children }: AppShellProps) {
	const [mobileOpen, setMobileOpen] = useState(false);

	return (
		<div className="flex h-screen overflow-hidden bg-background text-foreground">
			<Sheet open={mobileOpen} onOpenChange={setMobileOpen}>
				<SheetContent
					side="left"
					className="w-64 border-sidebar-border bg-sidebar p-0"
				>
					<SheetTitle className="sr-only">Navigation</SheetTitle>
					<Sidebar onNavClick={() => setMobileOpen(false)} />
				</SheetContent>
			</Sheet>

			<div className="flex min-w-0 flex-1 flex-col overflow-hidden">
				<Topbar onMenuClick={() => setMobileOpen(true)} />
				<main className="flex-1 overflow-y-auto p-4 sm:p-6 lg:p-8">
					<div className="mx-auto w-full max-w-7xl">{children}</div>
				</main>
			</div>
		</div>
	);
}
