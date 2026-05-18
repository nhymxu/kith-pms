import {
	BarChart3,
	BookOpen,
	Calendar,
	Gift,
	Heart,
	LayoutDashboard,
	Users,
} from "lucide-react";
import { NavLink } from "./nav-link";

export const NAV_ITEMS = [
	{ to: "/", icon: LayoutDashboard, label: "Dashboard" },
	{ to: "/people", icon: Users, label: "People" },
	{ to: "/journal", icon: BookOpen, label: "Journal" },
	{ to: "/gifts", icon: Gift, label: "Gifts" },
	{ to: "/dates", icon: Calendar, label: "Dates" },
	{ to: "/reminders", icon: Heart, label: "Reminders" },
	{ to: "/audit", icon: BarChart3, label: "Audit" },
] as const;

interface SidebarProps {
	onNavClick?: () => void;
}

export function Sidebar({ onNavClick }: SidebarProps) {
	return (
		<div className="flex h-full flex-col">
			<div className="flex h-14 items-center border-b border-zinc-200 px-5">
				<span className="text-[15px] font-semibold tracking-tight">Kith</span>
			</div>

			<nav className="flex-1 overflow-y-auto p-3 space-y-0.5">
				{NAV_ITEMS.map((item) => (
					<NavLink
						key={item.to}
						to={item.to}
						icon={item.icon}
						label={item.label}
						onClick={onNavClick}
					/>
				))}
			</nav>
		</div>
	);
}
