import { Link } from "@tanstack/react-router";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "#/components/ui/tooltip";
import { getAvatarUrl } from "#/endpoints/people";
import { formatDate } from "#/lib/format-datetime";
import type { Person } from "#/schemas/person";

export function PersonAvatarTooltip({ person }: { person: Person }) {
	const lastContact = person.last_contact_at
		? formatDate(person.last_contact_at)
		: "Never contacted";

	return (
		<Tooltip>
			<TooltipTrigger
				render={
					<Link
						to="/people/$personId"
						params={{ personId: String(person.id) }}
						className="size-16 rounded-full overflow-hidden shrink-0 bg-chip flex items-center justify-center text-lg font-medium text-chip-fg font-mono ring-2 ring-transparent hover:ring-accent transition-[box-shadow]"
					>
						{person.avatar_path ? (
							<img
								src={getAvatarUrl(person.id)}
								alt={person.name}
								className="size-full object-cover"
							/>
						) : (
							<span>{person.name.charAt(0).toUpperCase()}</span>
						)}
					</Link>
				}
			/>
			<TooltipContent>
				<p className="font-medium text-ink">
					{person.name}
					{person.nickname ? ` (${person.nickname})` : ""}
				</p>
				<p className="text-sub">Last contact: {lastContact}</p>
			</TooltipContent>
		</Tooltip>
	);
}
