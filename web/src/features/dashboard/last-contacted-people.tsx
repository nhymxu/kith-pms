import { Clock } from "lucide-react";
import { TooltipProvider } from "#/components/ui/tooltip";
import type { Person } from "#/schemas/person";
import { DashboardCard } from "./dashboard-card";
import { EmptyState } from "./empty-state";
import { PersonAvatarTooltip } from "./person-avatar-tooltip";

export function LastContactedPeople({
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
			title="Last Contact"
			subtitle="People you've recently connected with"
			icon={Clock}
			onRefresh={onRefresh}
			isRefreshing={isRefreshing}
		>
			{isLoading ? (
				<div className="flex flex-wrap gap-3">
					{["l1", "l2", "l3"].map((key) => (
						<div
							key={key}
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
					icon={Clock}
					title="No contact history yet"
					description="Log a journal entry or interaction to see recent contacts here."
				/>
			)}
		</DashboardCard>
	);
}
