import { mergeProps } from "@base-ui/react/merge-props";
import { useRender } from "@base-ui/react/use-render";
import { cva, type VariantProps } from "class-variance-authority";
import type * as React from "react";
import { cn } from "#/lib/utils";

const badgeVariants = cva(
	"inline-flex items-center justify-center rounded-base border-bw px-2 py-0.5 text-[11px] font-medium w-fit whitespace-nowrap shrink-0 [&>svg]:size-3 gap-1 [&>svg]:pointer-events-none overflow-hidden",
	{
		variants: {
			variant: {
				default: "border-chip-line bg-chip text-chip-fg",
				neutral: "border-line bg-panel text-sub",
				success: "border-success-line bg-success-bg text-success-fg",
				warning: "border-warning-line bg-warning-bg text-warning-fg",
				destructive: "border-danger-line bg-danger-bg text-danger-fg",
			},
		},
		defaultVariants: {
			variant: "default",
		},
	},
);

function Badge({
	className,
	variant,
	asChild = false,
	children,
	...props
}: React.ComponentProps<"span"> &
	VariantProps<typeof badgeVariants> & { asChild?: boolean }) {
	return useRender({
		render: asChild && children ? (children as React.ReactElement) : <span />,
		props: mergeProps(props, {
			children: asChild ? undefined : children,
			"data-slot": "badge",
			className: cn(badgeVariants({ variant }), className),
		}),
	});
}

export { Badge, badgeVariants };
