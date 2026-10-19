import { Menu, Search } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Button } from "#/components/ui/button";
import { NavLink } from "./nav-link";
import { NAV_ITEMS } from "./sidebar";
import { UserMenu } from "./user-menu";

interface TopbarProps {
	onMenuClick: () => void;
	/** Side-rail layout: search + user chip only, no horizontal nav or wordmark. */
	slim?: boolean;
}

function SearchButton() {
	const { t } = useTranslation("search");
	return (
		<button
			type="button"
			data-search-field
			onClick={() => {
				window.dispatchEvent(new CustomEvent("kith:open-command-palette"));
			}}
			className="inline-flex items-center gap-2 h-9 px-2.5 sm:px-3 sm:w-[200px] lg:w-[240px] rounded-md border-field-bw border-field-line bg-field text-sub hover:border-sub transition-colors shrink-0"
			aria-label={t("placeholder")}
		>
			<Search className="size-3.5 shrink-0" />
			<span className="hidden sm:block flex-1 text-left text-[13px] truncate">
				{t("placeholder")}
			</span>
			<kbd className="hidden sm:block font-mono text-[11px] opacity-75">⌘K</kbd>
		</button>
	);
}

export function Topbar({ onMenuClick, slim = false }: TopbarProps) {
	const { t } = useTranslation();
	return (
		<header className="sticky top-0 z-30 flex h-14 items-center gap-3 border-b border-line bg-topbar px-4 sm:px-6">
			<div className="flex items-center gap-3">
				<Button
					variant="ghost"
					size="icon"
					className="md:hidden"
					onClick={onMenuClick}
					aria-label={t("nav.open")}
				>
					<Menu className="size-5" />
				</Button>
				{!slim && (
					<span className="text-[15px] font-semibold tracking-tight">Kith</span>
				)}
			</div>

			{!slim && (
				<nav
					aria-label={t("nav.primary")}
					className="hidden md:flex items-center gap-1 text-[13px] flex-1"
				>
					{NAV_ITEMS.map((item) => (
						<NavLink
							key={item.to}
							to={item.to}
							icon={item.icon}
							label={t(item.labelKey)}
							variant="topbar"
						/>
					))}
				</nav>
			)}

			<div className="ml-auto flex items-center gap-3">
				<SearchButton />
				<UserMenu />
			</div>
		</header>
	);
}
