import * as React from "react"
import { cn } from "#/lib/utils"

function Input({ className, type, ...props }: React.ComponentProps<"input">) {
	return (
		<input
			type={type}
			data-slot="input"
			className={cn(
				"flex h-9 w-full rounded-md border border-zinc-200 bg-white px-3 py-2 text-[13px] text-foreground placeholder:text-zinc-400 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-600 focus-visible:ring-offset-1 disabled:cursor-not-allowed disabled:opacity-50 file:border-0 file:bg-transparent file:text-sm file:font-medium",
				className,
			)}
			{...props}
		/>
	)
}

export { Input }
