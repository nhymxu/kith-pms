import { Radio as RadioPrimitive } from "@base-ui/react/radio";
import { RadioGroup as RadioGroupPrimitive } from "@base-ui/react/radio-group";
import type * as React from "react";
import { cn } from "#/lib/utils";

type RadioGroupProps = Omit<
	React.ComponentProps<typeof RadioGroupPrimitive>,
	"onValueChange"
> & {
	onValueChange?: (value: string) => void;
};

function RadioGroup({ className, onValueChange, ...props }: RadioGroupProps) {
	return (
		<RadioGroupPrimitive
			data-slot="radio-group"
			className={cn("grid gap-2", className)}
			onValueChange={
				onValueChange ? (value) => onValueChange(String(value)) : undefined
			}
			{...props}
		/>
	);
}

function RadioGroupItem({
	className,
	...props
}: React.ComponentProps<typeof RadioPrimitive.Root>) {
	return (
		<RadioPrimitive.Root
			data-slot="radio-group-item"
			className={cn(
				"size-4 shrink-0 rounded-full border border-zinc-300 bg-white ring-offset-white focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-600 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 data-[checked]:border-indigo-600",
				className,
			)}
			{...props}
		>
			<RadioPrimitive.Indicator className="flex size-full items-center justify-center">
				<div className="size-2 rounded-full bg-indigo-600" />
			</RadioPrimitive.Indicator>
		</RadioPrimitive.Root>
	);
}

export { RadioGroup, RadioGroupItem };
