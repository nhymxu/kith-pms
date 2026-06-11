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
			className="flex items-center gap-1.5 rounded-full border border-zinc-200 bg-white px-2 py-0.5 hover:border-indigo-400 hover:bg-indigo-50 transition-colors"
		>
			<span className="size-5 rounded-full overflow-hidden shrink-0 bg-zinc-100 flex items-center justify-center text-[9px] font-medium text-zinc-600">
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
			<span className="text-[11px] text-zinc-700 leading-none">{display}</span>
		</Link>
	);
}
