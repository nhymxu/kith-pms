import { useNavigate } from "@tanstack/react-router"
import { LogOut, Settings, Shield, Tag, Users } from "lucide-react"
import { useAuth } from "#/lib/auth-context"
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuLabel,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "#/components/ui/dropdown-menu"

export function UserMenu() {
	const { user, logout } = useAuth()
	const navigate = useNavigate()

	const handleLogout = async () => {
		await logout()
	}

	const go = (path: string) => navigate({ to: path as "/" })

	const initials = user ? `U${user.id}` : "?"

	return (
		<DropdownMenu>
			<DropdownMenuTrigger asChild>
				<button
					type="button"
					className="inline-flex items-center gap-2 rounded-md px-2 py-1 text-[13px] hover:bg-zinc-100 transition-colors"
				>
					<span className="size-7 rounded-full bg-zinc-900 text-white text-[11px] font-medium grid place-items-center shrink-0">
						{initials}
					</span>
					<span className="max-w-[100px] truncate hidden sm:block">
						{user ? `User #${user.id}` : "Account"}
					</span>
				</button>
			</DropdownMenuTrigger>
			<DropdownMenuContent align="end" className="w-48">
				<DropdownMenuLabel>Settings</DropdownMenuLabel>
				<DropdownMenuSeparator />
				<DropdownMenuItem onSelect={() => go("/settings")}>
					<Tag className="size-4" />
					Setting
				</DropdownMenuItem>
				<DropdownMenuSeparator />
				<DropdownMenuItem onSelect={() => go("/me")}>
					<Settings className="size-4" />
					Self Profile
				</DropdownMenuItem>
				<DropdownMenuSeparator />
				<DropdownMenuItem onSelect={handleLogout}>
					<LogOut className="size-4" />
					Log out
				</DropdownMenuItem>
			</DropdownMenuContent>
		</DropdownMenu>
	)
}
