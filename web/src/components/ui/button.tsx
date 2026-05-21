import { Button as ButtonPrimitive } from "@base-ui/react/button";
import { cva, type VariantProps } from "class-variance-authority";
import * as React from "react";
import { cn } from "#/lib/utils";

const buttonVariants = cva(
	"inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-white transition-colors gap-2 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-600 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",
	{
		variants: {
			variant: {
				default: "bg-indigo-600 text-white hover:bg-indigo-700",
				noShadow: "bg-indigo-600 text-white hover:bg-indigo-700",
				neutral:
					"bg-zinc-100 text-zinc-900 border border-zinc-200 hover:bg-zinc-200",
				reverse: "bg-indigo-600 text-white hover:bg-indigo-700",
				ghost: "hover:bg-zinc-100 text-zinc-700 border border-transparent",
				outline:
					"border border-zinc-200 bg-white text-zinc-700 hover:bg-zinc-50",
				destructive:
					"bg-destructive text-destructive-foreground hover:bg-destructive/90",
			},
			size: {
				default: "h-9 px-4 py-2",
				sm: "h-8 px-3 text-[13px]",
				lg: "h-10 px-8",
				icon: "size-9",
			},
		},
		defaultVariants: {
			variant: "default",
			size: "default",
		},
	},
);

function Button({
	className,
	variant,
	size,
	asChild = false,
	children,
	...props
}: ButtonPrimitive.Props &
	VariantProps<typeof buttonVariants> & {
		asChild?: boolean;
	}) {
	const render =
		asChild && React.isValidElement(children) ? children : undefined;

	return (
		<ButtonPrimitive
			data-slot="button"
			className={cn(buttonVariants({ variant, size, className }))}
			nativeButton={!render}
			render={render}
			{...props}
		>
			{render ? undefined : children}
		</ButtonPrimitive>
	);
}

export { Button, buttonVariants };
