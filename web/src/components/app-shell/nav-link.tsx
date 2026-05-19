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
				className="relative inline-flex items-center gap-1.5 px-3 py-2 text-[13px] text-zinc-500 hover:text-zinc-900 transition-colors"
				activeProps={{
					className:
						"relative inline-flex items-center gap-1.5 px-3 py-2 text-[13px] text-zinc-900 after:absolute after:inset-x-3 after:-bottom-px after:h-[2px] after:bg-indigo-600",
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
			className="flex items-center gap-3 px-3 py-2 rounded-md text-[13px] text-zinc-700 hover:bg-zinc-100 transition-colors"
			activeProps={{
				className:
					"flex items-center gap-3 px-3 py-2 rounded-md text-[13px] bg-zinc-100 text-zinc-900 font-medium",
			}}
		>
			<Icon className="size-4 shrink-0" />
			<span>{label}</span>
		</Link>
	);
}
