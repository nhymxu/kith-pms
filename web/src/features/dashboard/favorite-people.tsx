import { Link } from "@tanstack/react-router";
import { Star } from "lucide-react";
import { getAvatarUrl } from "#/endpoints/people";
import type { Person } from "#/schemas/person";
import { DashboardCard } from "./dashboard-card";
import { EmptyState } from "./empty-state";

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
				<div className="space-y-px">
					{["f1", "f2", "f3"].map((key) => (
						<div key={key} className="h-12 bg-zinc-100 animate-pulse rounded" />
					))}
				</div>
			) : people.length ? (
				<div>
					{people.map((person) => (
						<Link
							key={person.id}
							to="/people/$personId"
							params={{ personId: String(person.id) }}
							className="flex items-center gap-2 py-2 border-b border-zinc-100 last:border-b-0 hover:bg-zinc-50 -mx-4 px-4 transition-colors"
						>
							<div className="size-7 rounded-full overflow-hidden shrink-0 bg-zinc-100 flex items-center justify-center text-[11px] font-medium text-zinc-700 font-mono">
								{person.avatar_path ? (
									<img
										src={getAvatarUrl(person.id)}
										alt={person.name}
										className="size-full object-cover"
									/>
								) : (
									<span>{person.name.charAt(0).toUpperCase()}</span>
								)}
							</div>
							<p className="truncate text-[13px] text-zinc-900">
								{person.name}
							</p>
						</Link>
					))}
				</div>
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
