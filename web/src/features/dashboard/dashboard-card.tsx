import { RefreshCw } from "lucide-react";
import type { ElementType, ReactNode } from "react";

export function DashboardCard({
	title,
	subtitle,
	icon: Icon,
	onRefresh,
	isRefreshing,
	children,
	className = "",
}: {
	title: string;
	subtitle?: string;
	icon?: ElementType;
	onRefresh?: () => void;
	isRefreshing?: boolean;
	children: ReactNode;
	className?: string;
}) {
	return (
		<div className={`border border-zinc-200 rounded-md bg-white ${className}`}>
			<div className="flex items-center justify-between px-4 py-3 border-b border-zinc-200">
				<div className="min-w-0">
					<p className="flex items-center gap-2 text-[13px] font-medium text-zinc-900">
						{Icon ? <Icon className="size-3.5 text-zinc-400 shrink-0" /> : null}
						{title}
					</p>
					{subtitle ? (
						<p className="text-[11px] text-zinc-500 mt-0.5">{subtitle}</p>
					) : null}
				</div>
				{onRefresh ? (
					<button
						type="button"
						className="size-6 shrink-0 flex items-center justify-center rounded text-zinc-400 hover:text-zinc-700 transition-colors"
						onClick={onRefresh}
						aria-label={`Refresh ${title}`}
					>
						<RefreshCw
							className={`size-3.5 ${isRefreshing ? "animate-spin" : ""}`}
						/>
					</button>
				) : null}
			</div>
			<div className="p-4">{children}</div>
		</div>
	);
}
