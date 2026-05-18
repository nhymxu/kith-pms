import { createFileRoute, Link } from "@tanstack/react-router"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "#/components/ui/card"
import { Tag, Link2, Shield } from "lucide-react"

export const Route = createFileRoute("/_authed/settings/")({
	component: SettingsPage,
})

const SETTINGS_LINKS = [
	{
		to: "/settings/labels" as const,
		icon: Tag,
		title: "Labels",
		description: "Create and manage labels for categorising people.",
	},
	{
		to: "/settings/relationship-types" as const,
		icon: Link2,
		title: "Relationship Types",
		description: "Define relationship types used when linking people.",
	},
	{
		to: "/settings/security" as const,
		icon: Shield,
		title: "Security",
		description: "Change your password and manage sessions.",
	},
]

function SettingsPage() {
	return (
		<div className="space-y-4 max-w-xl">
			<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">Settings</h1>
			<div className="space-y-3">
				{SETTINGS_LINKS.map(({ to, icon: Icon, title, description }) => (
					<Link key={to} to={to} className="block">
						<Card className="hover:border-zinc-300 transition-colors cursor-pointer">
							<CardHeader className="pb-2">
								<CardTitle className="text-[14px] font-medium flex items-center gap-2 text-zinc-900">
									<Icon className="size-4 text-zinc-400" />
									{title}
								</CardTitle>
							</CardHeader>
							<CardContent>
								<CardDescription>{description}</CardDescription>
							</CardContent>
						</Card>
					</Link>
				))}
			</div>
		</div>
	)
}
