import * as React from "react"
import { cn } from "#/lib/utils"

function Textarea({ className, ...props }: React.ComponentProps<"textarea">) {
	return (
		<textarea
			data-slot="textarea"
			className={cn(
				"flex min-h-[80px] w-full rounded-md border border-zinc-200 bg-white px-3 py-2 text-[13px] text-foreground placeholder:text-zinc-400 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-600 focus-visible:ring-offset-1 disabled:cursor-not-allowed disabled:opacity-50",
				className,
			)}
			{...props}
		/>
	)
}

export { Textarea }
