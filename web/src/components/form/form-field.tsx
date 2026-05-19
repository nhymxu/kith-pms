import type { AnyFieldApi } from "@tanstack/react-form";
import type { ComponentProps } from "react";
import { Input } from "#/components/ui/input";
import { Label } from "#/components/ui/label";
import { cn } from "#/lib/utils";

interface FormFieldProps extends ComponentProps<typeof Input> {
	field: AnyFieldApi;
	label: string;
	description?: string;
}

// Safely extract a displayable string from a TanStack Form / Zod error value.
function errorMessage(err: unknown): string {
	if (typeof err === "string") return err;
	if (
		err &&
		typeof err === "object" &&
		"message" in err &&
		typeof (err as { message: unknown }).message === "string"
	) {
		return (err as { message: string }).message;
	}
	return String(err);
}

// Generic form field: wires TanStack Form field API to neobrutalism Input + Label + Zod errors.
// Safe for string-typed fields; value is coerced to string for display only.
export function FormField({
	field,
	label,
	description,
	className,
	...inputProps
}: FormFieldProps) {
	const errors = field.state.meta.errors;
	const hasError = errors.length > 0;

	// Coerce value to string — this wrapper is intentionally text-only; callers with
	// number/boolean/Date fields should use dedicated field components.
	const displayValue =
		field.state.value === null || field.state.value === undefined
			? ""
			: typeof field.state.value === "string"
				? field.state.value
				: String(field.state.value);

	return (
		<div className="space-y-1.5">
			<Label
				htmlFor={field.name}
				className={cn(hasError && "text-destructive")}
			>
				{label}
			</Label>
			{description && (
				<p className="text-xs text-foreground/60">{description}</p>
			)}
			<Input
				id={field.name}
				name={field.name}
				value={displayValue}
				onBlur={field.handleBlur}
				onChange={(e) => field.handleChange(e.target.value)}
				aria-invalid={hasError}
				className={cn(
					hasError && "border-destructive ring-destructive",
					className,
				)}
				{...inputProps}
			/>
			{hasError && (
				<ul className="space-y-0.5">
					{errors.map((err, i) => (
						<li key={i} className="text-[11px] text-red-600 mt-1">
							{errorMessage(err)}
						</li>
					))}
				</ul>
			)}
		</div>
	);
}
