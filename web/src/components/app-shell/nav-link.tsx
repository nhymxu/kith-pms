import { Link } from "@tanstack/react-router";
import type { LucideIcon } from "lucide-react";
import { cn } from "#/lib/utils";

interface NavLinkProps {
	to: string;
	icon: LucideIcon;
	label: string;
	onClick?: () => void;
	variant?: "sidebar" | "topbar";
}

const navLinkStyles = {
	sidebar: {
		base: "text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
		active:
			"bg-sidebar-primary text-sidebar-primary-foreground font-heading shadow-shadow",
	},
	topbar: {
		base: "text-foreground hover:bg-secondary-background hover:text-foreground",
		active: "bg-main text-main-foreground font-heading shadow-shadow",
	},
} as const;

export function NavLink({
	to,
	icon: Icon,
	label,
	onClick,
	variant = "sidebar",
}: NavLinkProps) {
	const styles = navLinkStyles[variant];

	return (
		<Link
			to={to}
			onClick={onClick}
			className={cn(
				"flex items-center gap-3 rounded-base px-3.5 py-2.5 text-sm font-base transition-all",
				styles.base,
			)}
			activeProps={{ className: styles.active }}
		>
			<Icon className="size-4 shrink-0" />
			<span>{label}</span>
		</Link>
	);
}
