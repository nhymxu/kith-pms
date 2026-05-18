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
			className={`rounded-md px-2.5 py-1 text-[11px] font-medium transition-colors border ${
				active
					? "bg-zinc-900 text-white border-zinc-900"
					: "bg-white text-zinc-600 border-zinc-200 hover:bg-zinc-50"
			}`}
			onClick={onClick}
		>
			{label}
		</button>
	);
}
