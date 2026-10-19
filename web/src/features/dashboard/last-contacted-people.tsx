import { Clock } from "lucide-react";
import { useTranslation } from "react-i18next";
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
	const { t } = useTranslation("dashboard");
	return (
		<DashboardCard
			title={t("lastContacted.title")}
			subtitle={t("lastContacted.subtitle")}
			icon={Clock}
			onRefresh={onRefresh}
			isRefreshing={isRefreshing}
		>
			{isLoading ? (
				<div className="flex flex-wrap gap-3">
					{["l1", "l2", "l3"].map((key) => (
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
					icon={Clock}
					title={t("lastContacted.emptyTitle")}
					description={t("lastContacted.emptyDescription")}
				/>
			)}
		</DashboardCard>
	);
}
