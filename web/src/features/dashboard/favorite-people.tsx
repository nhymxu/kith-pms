import { Star } from "lucide-react";
import { TooltipProvider } from "#/components/ui/tooltip";
import type { Person } from "#/schemas/person";
import { DashboardCard } from "./dashboard-card";
import { EmptyState } from "./empty-state";
import { PersonAvatarTooltip } from "./person-avatar-tooltip";

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
							data-avatar
							className="size-16 rounded-full bg-chip animate-pulse"
						/>
					))}
				</div>
			) : people.length ? (
				<TooltipProvider>
					<div className="flex flex-wrap gap-3">
						{people.map((person) => (
							<PersonAvatarTooltip key={person.id} person={person} />
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
