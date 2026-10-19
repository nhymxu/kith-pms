import { Loader2 } from "lucide-react";
import type { ComponentProps } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "#/components/ui/button";
import { cn } from "#/lib/utils";

interface SubmitButtonProps extends ComponentProps<typeof Button> {
	isPending?: boolean;
	pendingLabel?: string;
}

export function SubmitButton({
	isPending = false,
	pendingLabel,
	children,
	disabled,
	className,
	...props
}: SubmitButtonProps) {
	const { t } = useTranslation();
	return (
		<Button
			type="submit"
			disabled={disabled || isPending}
			className={cn("gap-2", className)}
			{...props}
		>
			{isPending && <Loader2 className="size-4 animate-spin" />}
			{isPending ? (pendingLabel ?? t("saving")) : children}
		</Button>
	);
}
