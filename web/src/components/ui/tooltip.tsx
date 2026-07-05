import { Tooltip as TooltipPrimitive } from "@base-ui/react/tooltip";
import type * as React from "react";
import { cn } from "#/lib/utils";

function TooltipProvider({
	delay = 300,
	...props
}: React.ComponentProps<typeof TooltipPrimitive.Provider>) {
	return (
		<TooltipPrimitive.Provider
			data-slot="tooltip-provider"
			delay={delay}
			{...props}
		/>
	);
}

function Tooltip({
	...props
}: React.ComponentProps<typeof TooltipPrimitive.Root>) {
	return <TooltipPrimitive.Root data-slot="tooltip" {...props} />;
}

function TooltipTrigger({
	...props
}: React.ComponentProps<typeof TooltipPrimitive.Trigger>) {
	return <TooltipPrimitive.Trigger data-slot="tooltip-trigger" {...props} />;
}

function TooltipContent({
	className,
	sideOffset = 6,
	...props
}: React.ComponentProps<typeof TooltipPrimitive.Popup> & {
	sideOffset?: number;
}) {
	return (
		<TooltipPrimitive.Portal>
			<TooltipPrimitive.Positioner sideOffset={sideOffset}>
				<TooltipPrimitive.Popup
					data-slot="tooltip-content"
					className={cn(
						"z-50 rounded-md border border-zinc-200 bg-white px-3 py-1.5 text-[12px] text-zinc-700 shadow-lg outline-none data-[starting-style]:scale-95 data-[starting-style]:opacity-0 data-[ending-style]:scale-95 data-[ending-style]:opacity-0 transition-[scale,opacity]",
						className,
					)}
					{...props}
				/>
			</TooltipPrimitive.Positioner>
		</TooltipPrimitive.Portal>
	);
}

export { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger };
