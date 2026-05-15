import {
	BarChart3,
	BookOpen,
	Calendar,
	Gift,
	Heart,
	LayoutDashboard,
	Users,
} from "lucide-react"
import { NavLink } from "./nav-link"

const NAV_ITEMS = [
	{ to: "/", icon: LayoutDashboard, label: "Dashboard" },
	{ to: "/people", icon: Users, label: "People" },
	{ to: "/journal", icon: BookOpen, label: "Journal" },
	{ to: "/gifts", icon: Gift, label: "Gifts" },
	{ to: "/dates", icon: Calendar, label: "Dates" },
	{ to: "/reminders", icon: Heart, label: "Reminders" },
	{ to: "/audit", icon: BarChart3, label: "Audit" },
] as const

interface SidebarProps {
	onNavClick?: () => void
}

export function Sidebar({ onNavClick }: SidebarProps) {
	return (
		<div className="flex h-full flex-col">
			{/* Brand */}
			<div className="flex h-14 items-center border-b-2 border-border px-4">
				<span className="text-lg font-heading tracking-tight">Kith PMS</span>
			</div>

			{/* Nav */}
			<nav className="flex-1 overflow-y-auto p-3 space-y-1">
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
	)
}
