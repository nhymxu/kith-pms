import { type ReactNode, useState } from "react";
import { Sheet, SheetContent, SheetTitle } from "#/components/ui/sheet";
import { RequireAuth } from "#/lib/auth-context";
import { Footer } from "./footer";
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
		<div className="flex flex-col min-h-screen bg-background text-foreground">
			<Topbar onMenuClick={() => setMobileOpen(true)} />

			<Sheet open={mobileOpen} onOpenChange={setMobileOpen}>
				<SheetContent
					side="left"
					className="w-72 border-r border-zinc-200 bg-white p-0"
				>
					<SheetTitle className="sr-only">Navigation</SheetTitle>
					<Sidebar onNavClick={() => setMobileOpen(false)} />
				</SheetContent>
			</Sheet>

			<main className="flex-1 overflow-y-auto">
				<div className="mx-auto w-full max-w-[1440px] px-4 sm:px-6 py-6 lg:py-8">
					{children}
				</div>
			</main>
			<Footer />
		</div>
	);
}
