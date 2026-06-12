import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { LogOut, Settings, Tag } from "lucide-react";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuLabel,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "#/components/ui/dropdown-menu";
import { getMe } from "#/endpoints/me";
import { getAvatarUrl } from "#/endpoints/people";
import { useAuth } from "#/lib/auth-context";
import { keys } from "#/query-keys";

export function UserMenu() {
	const { user, logout } = useAuth();
	const navigate = useNavigate();

	const { data: profile } = useQuery({
		queryKey: keys.me.profile(),
		queryFn: getMe,
		retry: false,
		enabled: !!user,
	});

	const handleLogout = async () => {
		await logout();
	};

	const go = (path: string) => navigate({ to: path as "/" });

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
		<DropdownMenu>
			<DropdownMenuTrigger asChild>
				<button
					type="button"
					className="inline-flex items-center gap-2 rounded-md px-2 py-1 text-[13px] hover:bg-zinc-100 transition-colors"
				>
					<span className="size-7 rounded-full bg-zinc-900 text-white text-[11px] font-medium grid place-items-center shrink-0 overflow-hidden">
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
					<span className="max-w-[100px] truncate hidden sm:block">
						{displayName}
					</span>
				</button>
			</DropdownMenuTrigger>
			<DropdownMenuContent align="end" className="w-48">
				<DropdownMenuLabel>Settings</DropdownMenuLabel>
				<DropdownMenuSeparator />
				<DropdownMenuItem onClick={() => go("/settings")}>
					<Settings className="size-4" />
					Settings
				</DropdownMenuItem>
				<DropdownMenuSeparator />
				<DropdownMenuItem onClick={() => go("/me")}>
					<Tag className="size-4" />
					Self Profile
				</DropdownMenuItem>
				<DropdownMenuSeparator />
				<DropdownMenuItem onClick={handleLogout}>
					<LogOut className="size-4" />
					Log out
				</DropdownMenuItem>
			</DropdownMenuContent>
		</DropdownMenu>
	);
}
