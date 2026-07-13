import { Link } from "@tanstack/react-router";
import type { LucideIcon } from "lucide-react";

interface NavLinkProps {
	to: string;
	icon: LucideIcon;
	label: string;
	onClick?: () => void;
	variant?: "sidebar" | "topbar";
}

export function NavLink({
	to,
	icon: Icon,
	label,
	onClick,
	variant = "sidebar",
}: NavLinkProps) {
	if (variant === "topbar") {
		return (
			<Link
				to={to}
				onClick={onClick}
				className="relative inline-flex items-center gap-1.5 px-3 py-2 text-[13px] text-sub hover:text-ink transition-colors"
				activeProps={{
					className:
						"relative inline-flex items-center gap-1.5 px-3 py-2 text-[13px] text-nav-active-fg after:absolute after:inset-x-3 after:-bottom-px after:h-[2px] after:bg-nav-active",
				}}
			>
				<Icon className="size-3.5 shrink-0" />
				<span>{label}</span>
			</Link>
		);
	}

	return (
		<Link
			to={to}
			onClick={onClick}
			className="flex items-center gap-3 px-3 py-2 rounded-md text-[13px] text-sub hover:bg-chip transition-colors"
			activeProps={{
				className:
					"flex items-center gap-3 px-3 py-2 rounded-md text-[13px] bg-nav-active text-nav-active-fg font-medium",
			}}
		>
			<Icon className="size-4 shrink-0" />
			<span>{label}</span>
		</Link>
	);
}
