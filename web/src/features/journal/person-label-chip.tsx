import { Link } from "@tanstack/react-router";
import { Badge } from "#/components/ui/badge";
import { getAvatarUrl } from "#/endpoints/people";
import type { ActivityLabel, ActivityPerson } from "#/schemas/journal";

export function LabelChip({ label }: { label: ActivityLabel }) {
	return (
		<Badge variant="neutral" style={{ borderColor: label.color }}>
			{label.name}
		</Badge>
	);
}

export function PersonChip({ p }: { p: ActivityPerson }) {
	const hasAvatar = Boolean(p.avatar_path);
	const display = p.nickname ? p.nickname : p.name;
	return (
		<Link
			to="/people/$personId"
			params={{ personId: String(p.person_id) }}
			className="flex items-center gap-1.5 rounded-full border-bw border-line bg-panel px-2 py-0.5 hover:border-accent hover:bg-accent/10 transition-colors"
		>
			<span className="size-5 rounded-full overflow-hidden shrink-0 bg-chip flex items-center justify-center text-[9px] font-medium text-chip-fg">
				{hasAvatar ? (
					<img
						src={getAvatarUrl(p.person_id)}
						alt={p.name}
						className="size-full object-cover"
					/>
				) : (
					p.name.charAt(0).toUpperCase()
				)}
			</span>
			<span className="text-[11px] text-ink leading-none">{display}</span>
		</Link>
	);
}
