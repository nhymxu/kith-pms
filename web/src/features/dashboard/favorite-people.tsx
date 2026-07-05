import { Link } from "@tanstack/react-router";
import { Star } from "lucide-react";
import {
	Tooltip,
	TooltipContent,
	TooltipProvider,
	TooltipTrigger,
} from "#/components/ui/tooltip";
import { getAvatarUrl } from "#/endpoints/people";
import { formatDate } from "#/lib/format-datetime";
import type { Person } from "#/schemas/person";
import { DashboardCard } from "./dashboard-card";
import { EmptyState } from "./empty-state";

function FavoritePersonAvatar({ person }: { person: Person }) {
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
						className="size-16 rounded-full overflow-hidden shrink-0 bg-zinc-100 flex items-center justify-center text-lg font-medium text-zinc-700 font-mono ring-2 ring-transparent hover:ring-indigo-300 transition-[box-shadow]"
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
				<p className="font-medium text-zinc-900">
					{person.name}
					{person.nickname ? ` (${person.nickname})` : ""}
				</p>
				<p className="text-zinc-500">Last contact: {lastContact}</p>
			</TooltipContent>
		</Tooltip>
	);
}

export function FavoritePeople({
	people,
	isLoading,
	onRefresh,
	isRefreshing,
}: {
	people: Person[];
	isLoading: boolean;
	onRefresh: () => void;
	isRefreshing: boolean;
}) {
	return (
		<DashboardCard
			title="Favorites"
			subtitle="Quick access to your starred people"
			icon={Star}
			onRefresh={onRefresh}
			isRefreshing={isRefreshing}
		>
			{isLoading ? (
				<div className="flex flex-wrap gap-3">
					{["f1", "f2", "f3"].map((key) => (
						<div
							key={key}
							className="size-16 rounded-full bg-zinc-100 animate-pulse"
						/>
					))}
				</div>
			) : people.length ? (
				<TooltipProvider>
					<div className="flex flex-wrap gap-3">
						{people.map((person) => (
							<FavoritePersonAvatar key={person.id} person={person} />
						))}
					</div>
				</TooltipProvider>
			) : (
				<EmptyState
					icon={Star}
					title="No favorites yet"
					description="Star someone from their profile or the people list."
				/>
			)}
		</DashboardCard>
	);
}
