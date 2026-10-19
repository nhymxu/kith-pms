import { ApiError } from "#/lib/api-client";
import i18n from "#/lib/i18n";

// Localized text for an error shown to the user. Server errors map by `code`;
// an unmapped code keeps the server's English text, and a non-API failure
// (network drop, thrown Error) falls back to its own message.
export function apiErrorMessage(err: unknown, fallback?: string): string {
	if (err instanceof ApiError) {
		const statusKey = `errors:status.${err.status}` as const;
		if (err.code) {
			// Generic http_<status> codes share the status table.
			const keys = err.code.startsWith("http_")
				? [`errors:server.${err.code}`, statusKey]
				: `errors:server.${err.code}`;
			return i18n.t(keys, {
				...err.params,
				defaultValue: err.message,
			});
		}
		return i18n.t(statusKey, { defaultValue: err.message });
	}
	if (err instanceof TypeError) return i18n.t("errors:status.network");
	if (err instanceof Error && err.message) return err.message;
	return fallback ?? i18n.t("errors:server.internal");
}
