import { Menu } from "lucide-react";
import { Button } from "#/components/ui/button";
import { NavLink } from "./nav-link";
import { NAV_ITEMS } from "./sidebar";
import { UserMenu } from "./user-menu";

interface TopbarProps {
	onMenuClick: () => void;
}

export function Topbar({ onMenuClick }: TopbarProps) {
	return (
		<header className="sticky top-0 z-30 flex h-14 items-center gap-6 border-b border-zinc-200 bg-white px-4 sm:px-6">
			<div className="flex items-center gap-3">
				<Button
					variant="ghost"
					size="icon"
					className="md:hidden"
					onClick={onMenuClick}
					aria-label="Open navigation"
				>
					<Menu className="size-5" />
				</Button>
				<span className="text-[15px] font-semibold tracking-tight">Kith</span>
			</div>

			<nav
				aria-label="Primary navigation"
				className="hidden md:flex items-center gap-1 text-[13px] flex-1"
			>
				{NAV_ITEMS.map((item) => (
					<NavLink
						key={item.to}
						to={item.to}
						icon={item.icon}
						label={item.label}
						variant="topbar"
					/>
				))}
			</nav>

			<div className="ml-auto">
				<UserMenu />
			</div>
		</header>
	);
}
