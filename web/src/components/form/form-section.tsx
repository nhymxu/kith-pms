import type { ReactNode } from "react"
import { cn } from "#/lib/utils"

interface FormSectionProps {
	label: string
	description?: string
	children: ReactNode
	className?: string
}

export function FormSection({ label, description, children, className }: FormSectionProps) {
	return (
		<div className={cn("space-y-1.5", className)}>
			<div>
				<p className="text-sm font-heading">{label}</p>
				{description && (
					<p className="text-xs text-foreground/60 font-base mt-0.5">{description}</p>
				)}
			</div>
			{children}
		</div>
	)
}
