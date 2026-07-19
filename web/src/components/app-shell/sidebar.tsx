import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import {
	BarChart3,
	BookOpen,
	Calendar,
	Gift,
	Heart,
	LayoutDashboard,
	Network,
	StickyNote,
	Users,
} from "lucide-react";
import { getMe } from "#/endpoints/me";
import { getAvatarUrl } from "#/endpoints/people";
import { useAuth } from "#/lib/auth-context";
import { keys } from "#/query-keys";
import { NavLink } from "./nav-link";

export const NAV_ITEMS = [
	{ to: "/", icon: LayoutDashboard, label: "Dashboard" },
	{ to: "/people", icon: Users, label: "People" },
	{ to: "/network", icon: Network, label: "Network" },
	{ to: "/journal", icon: BookOpen, label: "Journal" },
	{ to: "/notes", icon: StickyNote, label: "Notes" },
	{ to: "/gifts", icon: Gift, label: "Gifts" },
	{ to: "/important-dates", icon: Calendar, label: "Dates" },
	{ to: "/reminders", icon: Heart, label: "Reminders" },
	{ to: "/audit", icon: BarChart3, label: "Audit" },
] as const;

interface SidebarProps {
	onNavClick?: () => void;
	/** Desktop rail footer (avatar/name/links). Omitted in the mobile drawer. */
	showFooter?: boolean;
}

function SidebarFooter() {
	const { user } = useAuth();

	const { data: profile } = useQuery({
		queryKey: keys.me.profile(),
		queryFn: getMe,
		retry: false,
		enabled: !!user,
	});

	const displayName = profile
		? profile.nickname || profile.name
		: user
			? `User #${user.id}`
			: "Account";

	const initials = profile
		? (profile.nickname || profile.name).charAt(0).toUpperCase()
		: user
			? `U${user.id}`
			: "?";

	return (
		<div className="mt-auto flex items-center gap-2.5 border-t border-line px-5 py-3.5 shrink-0">
			<span
				data-avatar
				className="size-8 rounded-full bg-accent text-accent-foreground text-[12px] font-medium grid place-items-center shrink-0 overflow-hidden"
			>
				{profile?.avatar_path ? (
					<img
						src={getAvatarUrl(profile.id)}
						alt={displayName}
						className="size-full object-cover"
					/>
				) : (
					initials
				)}
			</span>
			<div className="min-w-0 text-[12.5px] font-semibold truncate">
				<div className="truncate">{displayName}</div>
				<div className="text-[11px] font-normal text-sub truncate">
					<Link to="/me" className="hover:text-ink transition-colors">
						Self profile
					</Link>
					{" · "}
					<Link to="/settings" className="hover:text-ink transition-colors">
						Settings
					</Link>
				</div>
			</div>
		</div>
	);
}

export function Sidebar({ onNavClick, showFooter = false }: SidebarProps) {
	return (
		<div className="flex h-full flex-col bg-sidebar">
			<div className="flex h-14 items-center border-b border-line px-5 shrink-0">
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

			{showFooter && <SidebarFooter />}
		</div>
	);
}
