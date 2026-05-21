import { Dialog as SheetPrimitive } from "@base-ui/react/dialog";
import { X } from "lucide-react";
import * as React from "react";
import { cn } from "#/lib/utils";

function Sheet({ ...props }: React.ComponentProps<typeof SheetPrimitive.Root>) {
	return <SheetPrimitive.Root data-slot="sheet" {...props} />;
}

function SheetTrigger({
	asChild = false,
	children,
	...props
}: React.ComponentProps<typeof SheetPrimitive.Trigger> & {
	asChild?: boolean;
}) {
	const render =
		asChild && React.isValidElement(children) ? children : undefined;

	return (
		<SheetPrimitive.Trigger
			data-slot="sheet-trigger"
			render={render}
			{...props}
		>
			{render ? undefined : children}
		</SheetPrimitive.Trigger>
	);
}

function SheetClose({
	asChild = false,
	children,
	...props
}: React.ComponentProps<typeof SheetPrimitive.Close> & { asChild?: boolean }) {
	const render =
		asChild && React.isValidElement(children) ? children : undefined;

	return (
		<SheetPrimitive.Close data-slot="sheet-close" render={render} {...props}>
			{render ? undefined : children}
		</SheetPrimitive.Close>
	);
}

function SheetPortal({
	...props
}: React.ComponentProps<typeof SheetPrimitive.Portal>) {
	return <SheetPrimitive.Portal data-slot="sheet-portal" {...props} />;
}

function SheetOverlay({
	className,
	...props
}: React.ComponentProps<typeof SheetPrimitive.Backdrop>) {
	return (
		<SheetPrimitive.Backdrop
			data-slot="sheet-overlay"
			className={cn(
				"fixed inset-0 z-50 bg-black/40 data-[starting-style]:opacity-0 data-[ending-style]:opacity-0 transition-opacity",
				className,
			)}
			{...props}
		/>
	);
}

function SheetContent({
	className,
	children,
	side = "right",
	...props
}: React.ComponentProps<typeof SheetPrimitive.Popup> & {
	side?: "top" | "bottom" | "left" | "right";
}) {
	return (
		<SheetPortal>
			<SheetOverlay />
			<SheetPrimitive.Popup
				data-slot="sheet-content"
				className={cn(
					"bg-white fixed z-50 flex flex-col gap-4 border-zinc-200 shadow-lg transition ease-in-out duration-300",
					side === "right" &&
						"data-[ending-style]:translate-x-full data-[starting-style]:translate-x-full inset-y-0 right-0 h-full w-3/4 border-l sm:max-w-sm",
					side === "left" &&
						"data-[ending-style]:-translate-x-full data-[starting-style]:-translate-x-full inset-y-0 left-0 h-full w-3/4 border-r sm:max-w-sm",
					side === "top" &&
						"data-[ending-style]:-translate-y-full data-[starting-style]:-translate-y-full inset-x-0 top-0 h-auto border-b",
					side === "bottom" &&
						"data-[ending-style]:translate-y-full data-[starting-style]:translate-y-full inset-x-0 bottom-0 h-auto border-t",
					className,
				)}
				{...props}
			>
				{children}
				<SheetPrimitive.Close className="absolute right-4 top-4 rounded-md text-zinc-400 hover:text-zinc-700 focus:outline-none focus:ring-2 focus:ring-indigo-600 focus:ring-offset-2 disabled:pointer-events-none">
					<X className="h-4 w-4" />
					<span className="sr-only">Close</span>
				</SheetPrimitive.Close>
			</SheetPrimitive.Popup>
		</SheetPortal>
	);
}

function SheetHeader({ className, ...props }: React.ComponentProps<"div">) {
	return (
		<div
			data-slot="sheet-header"
			className={cn("flex flex-col gap-1.5 p-4", className)}
			{...props}
		/>
	);
}

function SheetFooter({ className, ...props }: React.ComponentProps<"div">) {
	return (
		<div
			data-slot="sheet-footer"
			className={cn("mt-auto flex flex-col gap-3 p-4", className)}
			{...props}
		/>
	);
}

function SheetTitle({
	className,
	...props
}: React.ComponentProps<typeof SheetPrimitive.Title>) {
	return (
		<SheetPrimitive.Title
			data-slot="sheet-title"
			className={cn("text-zinc-900 font-semibold text-[15px]", className)}
			{...props}
		/>
	);
}

function SheetDescription({
	className,
	...props
}: React.ComponentProps<typeof SheetPrimitive.Description>) {
	return (
		<SheetPrimitive.Description
			data-slot="sheet-description"
			className={cn("text-[13px] text-zinc-500", className)}
			{...props}
		/>
	);
}

export {
	Sheet,
	SheetClose,
	SheetContent,
	SheetDescription,
	SheetFooter,
	SheetHeader,
	SheetOverlay,
	SheetPortal,
	SheetTitle,
	SheetTrigger,
};
