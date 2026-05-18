import { RefreshCw } from "lucide-react";
import type { ElementType, ReactNode } from "react";
import { Button } from "#/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "#/components/ui/card";

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
		<Card
			className={`border-slate-200 bg-white/95 shadow-sm transition-all hover:-translate-y-0.5 hover:shadow-md ${className}`}
		>
			<CardHeader className="flex flex-row items-start justify-between gap-4 pb-3">
				<div className="min-w-0 space-y-1">
					<CardTitle className="flex items-center gap-2 text-sm font-heading text-slate-900">
						{Icon ? <Icon className="size-4 text-teal-600" /> : null}
						{title}
					</CardTitle>
					{subtitle ? (
						<p className="text-xs font-base text-slate-500">{subtitle}</p>
					) : null}
				</div>
				{onRefresh ? (
					<Button
						type="button"
						variant="neutral"
						size="icon"
						className="size-8 shrink-0 border-slate-200 bg-white text-slate-600 hover:bg-teal-50 hover:text-teal-700"
						onClick={onRefresh}
						aria-label={`Refresh ${title}`}
					>
						<RefreshCw
							className={`size-4 ${isRefreshing ? "animate-spin" : ""}`}
						/>
					</Button>
				) : null}
			</CardHeader>
			<CardContent>{children}</CardContent>
		</Card>
	);
}
