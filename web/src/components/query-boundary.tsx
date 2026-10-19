import { type ReactNode, Suspense } from "react";
import { useTranslation } from "react-i18next";

interface QueryBoundaryProps {
	children: ReactNode;
	fallback?: ReactNode;
}

function DefaultFallback() {
	const { t } = useTranslation();
	return (
		<div className="py-8 text-center text-[13px] text-sub">{t("loading")}</div>
	);
}

export function QueryBoundary({
	children,
	fallback = <DefaultFallback />,
}: QueryBoundaryProps) {
	return <Suspense fallback={fallback}>{children}</Suspense>;
}
