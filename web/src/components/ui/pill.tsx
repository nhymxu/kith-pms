import { cva, type VariantProps } from "class-variance-authority";
import type * as React from "react";
import { cn } from "#/lib/utils";

const pillVariants = cva(
	"inline-flex items-center gap-1 font-mono text-[10px] uppercase tracking-wide whitespace-nowrap",
	{
		variants: {
			variant: {
				success: "text-success-fg",
				warning: "text-warning-fg",
				danger: "text-danger-fg",
				accent: "text-accent-text",
				plain: "text-sub",
			},
			strike: {
				true: "line-through",
				false: "",
			},
		},
		defaultVariants: {
			variant: "plain",
			strike: false,
		},
	},
);

export type PillVariant = NonNullable<
	VariantProps<typeof pillVariants>["variant"]
>;

interface PillProps
	extends React.ComponentProps<"span">,
		VariantProps<typeof pillVariants> {}

function Pill({ className, variant, strike, ...props }: PillProps) {
	return (
		<span
			data-slot="pill"
			className={cn(pillVariants({ variant, strike }), className)}
			{...props}
		/>
	);
}

export { Pill, pillVariants };
