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
		<header className="flex h-16 items-center gap-4 border-b border-border/70 bg-background/80 px-4 backdrop-blur sm:px-6 lg:px-8">
			<div className="flex items-center gap-3">
				<Button
					variant="neutral"
					size="icon"
					className="md:hidden"
					onClick={onMenuClick}
					aria-label="Open navigation"
				>
					<Menu className="size-5" />
				</Button>
				<span className="whitespace-nowrap text-xl font-heading tracking-tight">
					Kith PMS
				</span>
			</div>

			<nav
				aria-label="Primary navigation"
				className="hidden min-w-0 flex-1 items-center gap-1 overflow-x-auto md:flex"
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

			<UserMenu />
		</header>
	);
}
