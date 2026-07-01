import { type ReactNode, Suspense } from "react";

interface QueryBoundaryProps {
	children: ReactNode;
	fallback?: ReactNode;
}

const defaultFallback = (
	<div className="py-8 text-center text-[13px] text-zinc-400">Loading…</div>
);

export function QueryBoundary({
	children,
	fallback = defaultFallback,
}: QueryBoundaryProps) {
	return <Suspense fallback={fallback}>{children}</Suspense>;
}
