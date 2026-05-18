import { Activity } from "lucide-react";
import {
	CartesianGrid,
	Line,
	LineChart,
	ResponsiveContainer,
	Tooltip,
	XAxis,
	YAxis,
} from "recharts";
import { DashboardCard } from "./dashboard-card";
import type { DashboardPulsePoint } from "./dashboard-data";
import { EmptyState } from "./empty-state";

export function RelationshipPulseChart({
	data,
	isLoading,
	onRefresh,
	isRefreshing,
}: {
	data: DashboardPulsePoint[];
	isLoading: boolean;
	onRefresh: () => void;
	isRefreshing: boolean;
}) {
	const hasActivity = data.some(
		(point) => point.entries > 0 || point.touches > 0,
	);

	return (
		<DashboardCard
			title="Relationship pulse"
			subtitle="Journal touches over the last 14 days"
			icon={Activity}
			onRefresh={onRefresh}
			isRefreshing={isRefreshing}
			className="xl:col-span-7"
		>
			{isLoading ? (
				<div className="h-72 rounded-base bg-slate-100 animate-pulse" />
			) : hasActivity ? (
				<div className="h-72">
					<ResponsiveContainer width="100%" height="100%">
						<LineChart
							data={data}
							margin={{ top: 8, right: 12, bottom: 0, left: -24 }}
						>
							<CartesianGrid
								stroke="#e2e8f0"
								strokeDasharray="4 4"
								vertical={false}
							/>
							<XAxis
								dataKey="date"
								tick={{ fill: "#64748b", fontSize: 11 }}
								tickLine={false}
								axisLine={false}
								minTickGap={24}
							/>
							<YAxis
								tick={{ fill: "#64748b", fontSize: 11 }}
								tickLine={false}
								axisLine={false}
								allowDecimals={false}
							/>
							<Tooltip
								content={<PulseTooltip />}
								cursor={{ stroke: "#14b8a6", strokeDasharray: "4 4" }}
							/>
							<Line
								type="monotone"
								dataKey="touches"
								name="Touches"
								stroke="#0f766e"
								strokeWidth={3}
								dot={false}
								activeDot={{ r: 5 }}
							/>
							<Line
								type="monotone"
								dataKey="entries"
								name="Entries"
								stroke="#94a3b8"
								strokeWidth={2}
								dot={false}
								activeDot={{ r: 4 }}
							/>
						</LineChart>
					</ResponsiveContainer>
				</div>
			) : (
				<EmptyState
					icon={Activity}
					title="No recent pulse yet"
					description="Add journal entries to see relationship activity trends."
				/>
			)}
		</DashboardCard>
	);
}

function PulseTooltip({
	active,
	payload,
	label,
}: {
	active?: boolean;
	payload?: Array<{ name?: string; value?: number; color?: string }>;
	label?: string;
}) {
	if (!active || !payload?.length) return null;

	return (
		<div className="rounded-base border-2 border-slate-200 bg-white p-3 shadow-sm">
			<p className="text-xs font-heading text-slate-900">{label}</p>
			{payload.map((entry) => (
				<p
					key={entry.name}
					className="mt-1 text-xs font-base"
					style={{ color: entry.color }}
				>
					{entry.name}: {entry.value ?? 0}
				</p>
			))}
		</div>
	);
}
