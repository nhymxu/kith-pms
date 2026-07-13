// Values reference the theme's chart/surface tokens (see styles.css) instead of
// hardcoded hex so the dashboard chart recolors across all 6 themes.
export const CHART_COLORS = {
	primary: "var(--chart-1)",
	secondary: "var(--chart-2)",
	grid: "var(--line)",
	axis: "var(--sub)",
	tooltipBg: "var(--panel)",
	tooltipBorder: "var(--line)",
} as const;
