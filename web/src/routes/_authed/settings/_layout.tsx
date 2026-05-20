import {
	createFileRoute,
	Link,
	Outlet,
	useRouterState,
} from "@tanstack/react-router";
import { Link2, Settings, Shield, Tag } from "lucide-react";

export const Route = createFileRoute("/_authed/settings/_layout")({
	component: SettingsLayout,
});

const NAV_ITEMS = [
	{ to: "/settings/general" as const, icon: Settings, label: "General" },
	{ to: "/settings/labels" as const, icon: Tag, label: "Labels" },
	{
		to: "/settings/relationship-types" as const,
		icon: Link2,
		label: "Relationship Types",
	},
	{ to: "/settings/security" as const, icon: Shield, label: "Security" },
];

function SettingsLayout() {
	const pathname = useRouterState({ select: (s) => s.location.pathname });

	return (
		<div className="flex gap-8 min-h-[60vh]">
			{/* Sidebar */}
			<nav className="w-48 shrink-0">
				<p className="text-[11px] font-semibold uppercase tracking-widest text-zinc-400 mb-3">
					Settings
				</p>
				<ul className="space-y-0.5">
					{NAV_ITEMS.map(({ to, icon: Icon, label }) => {
						const active = pathname === to || pathname.startsWith(`${to}/`);
						return (
							<li key={to}>
								<Link
									to={to}
									className={`flex items-center gap-2 px-3 py-2 rounded-md text-[13px] transition-colors ${
										active
											? "bg-indigo-50 text-indigo-700 font-medium"
											: "text-zinc-600 hover:bg-zinc-100 hover:text-zinc-900"
									}`}
								>
									<Icon className="size-4 shrink-0" />
									{label}
								</Link>
							</li>
						);
					})}
				</ul>
			</nav>

			{/* Content panel */}
			<div className="flex-1 min-w-0">
				<Outlet />
			</div>
		</div>
	);
}
