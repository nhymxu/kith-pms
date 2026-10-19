import type { ComponentProps, ReactElement } from "react";
import { Trans } from "react-i18next";

// i18n escapeValue is off (React escapes text), but Trans parses the
// interpolated string for tags, so user data must be escaped first.
export function SafeTrans(props: ComponentProps<typeof Trans>): ReactElement {
	return (
		<Trans
			tOptions={{ interpolation: { escapeValue: true } }}
			shouldUnescape
			{...props}
		/>
	);
}
