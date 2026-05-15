import { Link } from "@tanstack/react-router"
import { cn } from "#/lib/utils"
import type { LucideIcon } from "lucide-react"

interface NavLinkProps {
	to: string
	icon: LucideIcon
	label: string
	onClick?: () => void
}

export function NavLink({ to, icon: Icon, label, onClick }: NavLinkProps) {
	return (
		<Link
			to={to}
			onClick={onClick}
			className={cn(
				"flex items-center gap-3 px-3 py-2 rounded-base text-sm font-base border-2 border-transparent transition-all",
				"hover:border-border hover:bg-main hover:text-main-foreground hover:shadow-shadow hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-none",
			)}
			activeProps={{ className: "border-border bg-main text-main-foreground font-heading" }}
		>
			<Icon className="size-4 shrink-0" />
			<span>{label}</span>
		</Link>
	)
}
