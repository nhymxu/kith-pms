import { useQuery } from "@tanstack/react-query";
import { type ReactNode, useState } from "react";
import { Sheet, SheetContent, SheetTitle } from "#/components/ui/sheet";
import { getSettings } from "#/endpoints/settings";
import { CommandPalette } from "#/features/search/command-palette";
import { RequireAuth } from "#/lib/auth-context";
import { getNavLayout, type NavLayout } from "#/lib/nav-layout";
import { Footer } from "./footer";
import { Sidebar } from "./sidebar";
import { Topbar } from "./topbar";

interface AppShellProps {
	children: ReactNode;
}

export function AppShell({ children }: AppShellProps) {
	return (
		<RequireAuth>
			<CommandPalette />
			<AppShellInner>{children}</AppShellInner>
		</RequireAuth>
	);
}

// Shares the ["settings"] cache entry with the appearance picker and Footer, so
// switching nav_layout applies on next render (no reload): the picker's mutation
// calls queryClient.setQueryData(["settings"], saved), which notifies this
// subscriber immediately. Falls back to the localStorage-cached value (P01)
// until the query resolves.
function useNavLayoutMode(): NavLayout {
	const { data } = useQuery({
		queryKey: ["settings"],
		queryFn: getSettings,
		staleTime: 5 * 60 * 1000,
	});
	return data?.nav_layout ?? getNavLayout();
}

const mainContentClass =
	"mx-auto w-full max-w-[1440px] px-4 sm:px-6 py-6 lg:py-8";

function AppShellInner({ children }: AppShellProps) {
	const [mobileOpen, setMobileOpen] = useState(false);
	const layout = useNavLayoutMode();

	const mobileDrawer = (
		<Sheet open={mobileOpen} onOpenChange={setMobileOpen}>
			<SheetContent
				side="left"
				className="w-72 border-r border-line bg-sidebar p-0"
			>
				<SheetTitle className="sr-only">Navigation</SheetTitle>
				<Sidebar onNavClick={() => setMobileOpen(false)} />
			</SheetContent>
		</Sheet>
	);

	if (layout === "side") {
		return (
			<div className="flex min-h-screen bg-bg text-foreground">
				<aside className="hidden md:flex md:sticky md:top-0 md:h-screen md:self-start w-[212px] shrink-0 border-r border-line">
					<Sidebar showFooter />
				</aside>

				{mobileDrawer}

				<div className="flex flex-1 min-w-0 flex-col">
					<Topbar onMenuClick={() => setMobileOpen(true)} slim />
					<main className="flex-1 overflow-y-auto">
						<div className={mainContentClass}>{children}</div>
					</main>
					<Footer />
				</div>
			</div>
		);
	}

	return (
		<div className="flex flex-col min-h-screen bg-bg text-foreground">
			<Topbar onMenuClick={() => setMobileOpen(true)} />

			{mobileDrawer}

			<main className="flex-1 overflow-y-auto">
				<div className={mainContentClass}>{children}</div>
			</main>
			<Footer />
		</div>
	);
}
