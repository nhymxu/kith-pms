import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"
import type * as React from "react"
import { cn } from "#/lib/utils"

const buttonVariants = cva(
	"inline-flex items-center justify-center whitespace-nowrap rounded-base text-sm font-base ring-offset-white transition-all gap-2 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0 focus-visible:outline-hidden focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",
	{
		variants: {
			variant: {
				default:
					"text-main-foreground bg-main border border-main shadow-shadow hover:bg-main/90 hover:-translate-y-0.5",
				noShadow: "text-main-foreground bg-main border border-main hover:bg-main/90",
				neutral:
					"bg-secondary-background text-foreground border border-border shadow-shadow hover:border-ring/40 hover:bg-white",
				reverse:
					"text-main-foreground bg-main border border-main hover:-translate-y-0.5 hover:shadow-shadow",
				ghost: "hover:bg-main/10 hover:text-foreground border border-transparent",
				destructive:
					"bg-destructive text-destructive-foreground border border-destructive shadow-shadow hover:bg-destructive/90",
			},
			size: {
				default: "h-10 px-4 py-2",
				sm: "h-9 px-3",
				lg: "h-11 px-8",
				icon: "size-10",
			},
		},
		defaultVariants: {
			variant: "default",
			size: "default",
		},
	},
)

function Button({
	className,
	variant,
	size,
	asChild = false,
	...props
}: React.ComponentProps<"button"> &
	VariantProps<typeof buttonVariants> & {
		asChild?: boolean
	}) {
	const Comp = asChild ? Slot : "button"

	return (
		<Comp
			data-slot="button"
			className={cn(buttonVariants({ variant, size, className }))}
			{...props}
		/>
	)
}

export { Button, buttonVariants }
