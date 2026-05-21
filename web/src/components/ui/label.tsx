import type * as React from "react";
import { cn } from "#/lib/utils";

function Label({ className, ...props }: React.ComponentProps<"label">) {
	return (
		// biome-ignore lint/a11y/noLabelWithoutControl: Labels are associated at call sites through htmlFor or wrapped controls.
		<label
			data-slot="label"
			className={cn(
				"text-[12px] font-medium text-zinc-700 leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70",
				className,
			)}
			{...props}
		/>
	);
}

export { Label };
