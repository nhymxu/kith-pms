import { Star } from "lucide-react";
import { useTranslation } from "react-i18next";
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
	const { t } = useTranslation("dashboard");
	return (
		<DashboardCard
			title={t("favorites.title")}
			subtitle={t("favorites.subtitle")}
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
					title={t("favorites.emptyTitle")}
					description={t("favorites.emptyDescription")}
				/>
			)}
		</DashboardCard>
	);
}
