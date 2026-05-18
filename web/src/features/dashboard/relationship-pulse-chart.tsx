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
import { CHART_COLORS } from "./chart-theme";
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
		>
			{isLoading ? (
				<div className="h-60 rounded bg-zinc-100 animate-pulse" />
			) : hasActivity ? (
				<div className="h-60">
					<ResponsiveContainer width="100%" height="100%">
						<LineChart
							data={data}
							margin={{ top: 8, right: 12, bottom: 0, left: -24 }}
						>
							<CartesianGrid
								stroke={CHART_COLORS.grid}
								strokeDasharray="4 4"
								vertical={false}
							/>
							<XAxis
								dataKey="date"
								tick={{ fill: CHART_COLORS.axis, fontSize: 11 }}
								tickLine={false}
								axisLine={false}
								minTickGap={24}
							/>
							<YAxis
								tick={{ fill: CHART_COLORS.axis, fontSize: 11 }}
								tickLine={false}
								axisLine={false}
								allowDecimals={false}
							/>
							<Tooltip
								content={<PulseTooltip />}
								cursor={{ stroke: CHART_COLORS.primary, strokeDasharray: "4 4" }}
							/>
							<Line
								type="monotone"
								dataKey="touches"
								name="Touches"
								stroke={CHART_COLORS.primary}
								strokeWidth={2}
								dot={false}
								activeDot={{ r: 4, fill: CHART_COLORS.primary, stroke: '#fff', strokeWidth: 2 }}
							/>
							<Line
								type="monotone"
								dataKey="entries"
								name="Entries"
								stroke={CHART_COLORS.secondary}
								strokeWidth={1.5}
								dot={false}
								activeDot={{ r: 3, fill: CHART_COLORS.secondary, stroke: '#fff', strokeWidth: 2 }}
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
		<div className="border border-zinc-200 bg-white p-2 text-[11px] font-mono rounded-md">
			<p className="text-zinc-900 mb-1">{label}</p>
			{payload.map((entry) => (
				<p key={entry.name} style={{ color: entry.color }}>
					{entry.name}: {entry.value ?? 0}
				</p>
			))}
		</div>
	);
}
