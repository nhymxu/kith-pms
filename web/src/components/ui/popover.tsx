import { Popover as PopoverPrimitive } from "@base-ui/react/popover";
import * as React from "react";
import { cn } from "#/lib/utils";

function Popover({
	...props
}: React.ComponentProps<typeof PopoverPrimitive.Root>) {
	return <PopoverPrimitive.Root data-slot="popover" {...props} />;
}

function PopoverTrigger({
	asChild = false,
	children,
	...props
}: React.ComponentProps<typeof PopoverPrimitive.Trigger> & {
	asChild?: boolean;
}) {
	const render =
		asChild && React.isValidElement(children) ? children : undefined;

	return (
		<PopoverPrimitive.Trigger
			data-slot="popover-trigger"
			render={render}
			{...props}
		>
			{render ? undefined : children}
		</PopoverPrimitive.Trigger>
	);
}

function PopoverContent({
	className,
	sideOffset = 6,
	...props
}: React.ComponentProps<typeof PopoverPrimitive.Popup> & {
	sideOffset?: number;
}) {
	return (
		<PopoverPrimitive.Portal>
			<PopoverPrimitive.Positioner sideOffset={sideOffset}>
				<PopoverPrimitive.Popup
					data-slot="popover-content"
					className={cn(
						"z-50 rounded-md border border-zinc-200 bg-white p-3 shadow-lg outline-none data-[starting-style]:scale-95 data-[starting-style]:opacity-0 data-[ending-style]:scale-95 data-[ending-style]:opacity-0 transition-[scale,opacity]",
						className,
					)}
					{...props}
				/>
			</PopoverPrimitive.Positioner>
		</PopoverPrimitive.Portal>
	);
}

export { Popover, PopoverContent, PopoverTrigger };
