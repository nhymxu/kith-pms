import { Loader2 } from "lucide-react";
import type { ComponentProps } from "react";
import { Button } from "#/components/ui/button";
import { cn } from "#/lib/utils";

interface SubmitButtonProps extends ComponentProps<typeof Button> {
	isPending?: boolean;
	pendingLabel?: string;
}

export function SubmitButton({
	isPending = false,
	pendingLabel = "Saving…",
	children,
	disabled,
	className,
	...props
}: SubmitButtonProps) {
	return (
		<Button
			type="submit"
			disabled={disabled || isPending}
			className={cn("gap-2", className)}
			{...props}
		>
			{isPending && <Loader2 className="size-4 animate-spin" />}
			{isPending ? pendingLabel : children}
		</Button>
	);
}
