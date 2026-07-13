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
		<div
			className={`border-bw border-line rounded-base bg-panel shadow-shadow ${className}`}
		>
			<div className="flex items-center justify-between px-4 py-3 border-b border-line">
				<div className="min-w-0">
					<p className="flex items-center gap-2 text-[13px] font-medium text-ink">
						{Icon ? <Icon className="size-3.5 text-sub shrink-0" /> : null}
						{title}
					</p>
					{subtitle ? (
						<p className="text-[11px] text-sub mt-0.5">{subtitle}</p>
					) : null}
				</div>
				{onRefresh ? (
					<button
						type="button"
						className="size-6 shrink-0 flex items-center justify-center rounded text-sub hover:text-ink transition-colors"
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
