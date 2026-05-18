import { cva, type VariantProps } from "class-variance-authority"
import * as React from "react"
import { cn } from "#/lib/utils"

const alertVariants = cva(
	"relative w-full rounded-md border px-4 py-3 text-[13px] grid has-[>svg]:grid-cols-[calc(var(--spacing)*4)_1fr] grid-cols-[0_1fr] has-[>svg]:gap-x-3 gap-y-0.5 items-start [&>svg]:size-4 [&>svg]:translate-y-0.5 [&>svg]:text-current",
	{
		variants: {
			variant: {
				default: "border-zinc-200 bg-white text-zinc-700",
				destructive: "border-red-200 bg-red-50 text-red-700",
				warning: "border-amber-200 bg-amber-50 text-amber-700",
			},
		},
		defaultVariants: { variant: "default" },
	},
)

function Alert({ className, variant, ...props }: React.ComponentProps<"div"> & VariantProps<typeof alertVariants>) {
	return <div data-slot="alert" role="alert" className={cn(alertVariants({ variant }), className)} {...props} />
}

function AlertTitle({ className, ...props }: React.ComponentProps<"div">) {
	return <div data-slot="alert-title" className={cn("col-start-2 line-clamp-1 min-h-4 font-semibold tracking-tight", className)} {...props} />
}

function AlertDescription({ className, ...props }: React.ComponentProps<"div">) {
	return (
		<div
			data-slot="alert-description"
			className={cn("col-start-2 grid justify-items-start gap-1 text-[13px] [&_p]:leading-relaxed", className)}
			{...props}
		/>
	)
}

export { Alert, AlertTitle, AlertDescription }
