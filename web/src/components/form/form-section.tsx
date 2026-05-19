import type { ReactNode } from "react";
import { cn } from "#/lib/utils";

interface FormSectionProps {
	label: string;
	description?: string;
	children: ReactNode;
	className?: string;
}

export function FormSection({
	label,
	description,
	children,
	className,
}: FormSectionProps) {
	return (
		<div className={cn("space-y-1.5", className)}>
			<div>
				<p className="text-[13px] font-semibold text-zinc-900 pb-2 border-b border-zinc-200">
					{label}
				</p>
				{description && (
					<p className="text-[11px] text-zinc-500 mt-1">{description}</p>
				)}
			</div>
			{children}
		</div>
	);
}
