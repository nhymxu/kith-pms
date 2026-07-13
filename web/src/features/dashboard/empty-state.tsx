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
		<div className="py-8 text-center">
			<Icon className="size-5 text-sub mx-auto" />
			<p className="text-[13px] text-ink mt-2">{title}</p>
			<p className="text-[11px] text-sub mt-1">{description}</p>
		</div>
	);
}
