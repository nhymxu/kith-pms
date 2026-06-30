export function formatPersonName(
	name: string,
	nickname?: string | null,
): string {
	return nickname ? `${name} (${nickname})` : name;
}
