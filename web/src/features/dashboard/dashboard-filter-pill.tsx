import { Button } from "#/components/ui/button";

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
		<Button
			type="button"
			variant={active ? "default" : "neutral"}
			size="sm"
			className={
				active
					? "h-8 border-teal-700 bg-teal-600 px-3 text-xs text-white shadow-none hover:bg-teal-700"
					: "h-8 border-slate-200 bg-white px-3 text-xs text-slate-600 shadow-none hover:bg-teal-50 hover:text-teal-700"
			}
			onClick={onClick}
		>
			{label}
		</Button>
	);
}
