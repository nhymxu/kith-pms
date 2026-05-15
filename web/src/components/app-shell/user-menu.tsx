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
import { Button } from "#/components/ui/button"

export function UserMenu() {
	const { user, logout } = useAuth()
	const navigate = useNavigate()

	const handleLogout = async () => {
		await logout()
	}

	// Routes for settings pages will be added in Phase 5; cast to string to avoid route-tree errors.
	const go = (path: string) => navigate({ to: path as "/" })

	return (
		<DropdownMenu>
			<DropdownMenuTrigger asChild>
				<Button variant="neutral" size="sm" className="gap-2">
					<span className="max-w-[120px] truncate text-xs font-heading">
						{user ? `User #${user.id}` : "Account"}
					</span>
				</Button>
			</DropdownMenuTrigger>
			<DropdownMenuContent align="end" className="w-48">
				<DropdownMenuLabel>Settings</DropdownMenuLabel>
				<DropdownMenuSeparator />
				<DropdownMenuItem onSelect={() => go("/labels")}>
					<Tag className="size-4" />
					Labels
				</DropdownMenuItem>
				<DropdownMenuItem onSelect={() => go("/relationship-types")}>
					<Users className="size-4" />
					Relationship Types
				</DropdownMenuItem>
				<DropdownMenuSeparator />
				<DropdownMenuItem onSelect={() => go("/me")}>
					<Settings className="size-4" />
					Self Profile
				</DropdownMenuItem>
				<DropdownMenuItem onSelect={() => go("/security")}>
					<Shield className="size-4" />
					Security
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
