import { Menu } from "lucide-react"
import { Button } from "#/components/ui/button"
import { UserMenu } from "./user-menu"

interface TopbarProps {
	onMenuClick: () => void
}

export function Topbar({ onMenuClick }: TopbarProps) {
	return (
		<header className="flex h-14 items-center justify-between border-b-2 border-border bg-background px-4 gap-4">
			{/* Mobile menu trigger */}
			<Button
				variant="neutral"
				size="icon"
				className="md:hidden"
				onClick={onMenuClick}
				aria-label="Open navigation"
			>
				<Menu className="size-5" />
			</Button>

			{/* Spacer — breadcrumbs will go here in Phase 5 */}
			<div className="flex-1" />

			<UserMenu />
		</header>
	)
}
