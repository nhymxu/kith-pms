import {
	createFileRoute,
	Link,
	Outlet,
	useRouterState,
} from "@tanstack/react-router";
import { BookOpen, Link2, Palette, Settings, Shield, Tag } from "lucide-react";
import { useTranslation } from "react-i18next";

export const Route = createFileRoute("/_authed/settings/_layout")({
	component: SettingsLayout,
});

const NAV_ITEMS = [
	{ to: "/settings/general" as const, icon: Settings, labelKey: "nav.general" },
	{
		to: "/settings/appearance" as const,
		icon: Palette,
		labelKey: "nav.appearance",
	},
	{
		to: "/settings/people-labels" as const,
		icon: Tag,
		labelKey: "nav.peopleLabels",
	},
	{
		to: "/settings/journal-labels" as const,
		icon: BookOpen,
		labelKey: "nav.journalLabels",
	},
	{
		to: "/settings/relationship-types" as const,
		icon: Link2,
		labelKey: "nav.relationshipTypes",
	},
	{
		to: "/settings/security" as const,
		icon: Shield,
		labelKey: "nav.security",
	},
] as const;

function SettingsLayout() {
	const { t } = useTranslation("settings");
	const pathname = useRouterState({ select: (s) => s.location.pathname });

	return (
		<div className="flex gap-8 min-h-[60vh]">
			{/* Sidebar */}
			<nav className="w-48 shrink-0">
				<p className="text-[11px] font-semibold uppercase tracking-widest text-sub mb-3">
					{t("nav.title")}
				</p>
				<ul className="space-y-0.5">
					{NAV_ITEMS.map(({ to, icon: Icon, labelKey }) => {
						const active = pathname === to || pathname.startsWith(`${to}/`);
						return (
							<li key={to}>
								<Link
									to={to}
									className={`flex items-center gap-2 px-3 py-2 rounded-base text-[13px] transition-colors ${
										active
											? "bg-nav-active text-nav-active-fg font-medium"
											: "text-sub hover:bg-chip hover:text-ink"
									}`}
								>
									<Icon className="size-4 shrink-0" />
									{t(labelKey)}
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
