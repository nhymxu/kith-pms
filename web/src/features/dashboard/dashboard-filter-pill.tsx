export function DashboardFilterPill({
	label,
	active,
	onClick,
}: {
	label: string;
	active: boolean;
	onClick: () => void;
}) {
	return (
		<button
			type="button"
			className={`rounded-md px-2.5 py-1 text-[11px] font-medium transition-colors border-bw ${
				active
					? "bg-accent text-accent-foreground border-accent"
					: "bg-panel text-sub border-line hover:bg-chip"
			}`}
			onClick={onClick}
		>
			{label}
		</button>
	);
}
