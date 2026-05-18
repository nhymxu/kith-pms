import type { ElementType } from "react";

export function EmptyState({
	icon: Icon,
	title,
	description,
}: {
	icon: ElementType;
	title: string;
	description: string;
}) {
	return (
		<div className="rounded-base border-2 border-dashed border-slate-200 bg-slate-50/70 p-6 text-center">
			<Icon className="mx-auto size-8 text-slate-400" />
			<p className="mt-3 text-sm font-heading text-slate-800">{title}</p>
			<p className="mt-1 text-sm font-base text-slate-500">{description}</p>
		</div>
	);
}
