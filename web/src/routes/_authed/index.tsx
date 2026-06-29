import {
	type QueryClient,
	useQuery,
	useQueryClient,
} from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useMemo, useState } from "react";
import { listAudit } from "#/endpoints/audit";
import { listGifts } from "#/endpoints/gifts";
import { listUpcomingDates } from "#/endpoints/important-dates";
import { listJournal } from "#/endpoints/journal";
import { getMe } from "#/endpoints/me";
import { listPeople } from "#/endpoints/people";
import { listReminders } from "#/endpoints/reminders";
import { ActionQueue } from "#/features/dashboard/action-queue";
import {
	buildDashboardViewModel,
	type DashboardSummaryCard,
} from "#/features/dashboard/dashboard-data";
import { RecentRelationshipActivity } from "#/features/dashboard/recent-relationship-activity";
import { RelationshipPulseChart } from "#/features/dashboard/relationship-pulse-chart";
import { SummaryCards } from "#/features/dashboard/summary-cards";
import { UpcomingMoments } from "#/features/dashboard/upcoming-moments";
import { keys } from "#/query-keys";

export const Route = createFileRoute("/_authed/")({
	component: DashboardPage,
});

function DashboardPage() {
	const queryClient = useQueryClient();
	const [refreshingId, setRefreshingId] = useState<string>();
	const people = useQuery({
		queryKey: keys.people.list({ page_size: 25 }),
		queryFn: () => listPeople({ page_size: 25 }),
	});
	const journal = useQuery({
		queryKey: keys.journal.list({ page_size: 25 }),
		queryFn: () => listJournal({ page_size: 25 }),
	});
	const reminders = useQuery({
		queryKey: keys.reminders.list({ status: "all" }),
		queryFn: () => listReminders({ status: "all" }),
	});
	const dates = useQuery({
		queryKey: keys.dates.upcoming(),
		queryFn: () => listUpcomingDates(30),
	});
	const gifts = useQuery({
		queryKey: keys.gifts.list({ page_size: 25 }),
		queryFn: () => listGifts({ page_size: 25 }),
	});
	const audit = useQuery({
		queryKey: keys.audit.list({ page: 1 }),
		queryFn: () => listAudit({ page: 1 }),
	});
	const me = useQuery({
		queryKey: keys.me.profile(),
		queryFn: getMe,
		retry: false,
	});
	const viewModel = useMemo(
		() =>
			buildDashboardViewModel({
				people: people.data,
				journal: journal.data,
				reminders: reminders.data,
				dates: dates.data,
				gifts: gifts.data,
				audit: audit.data?.data,
				me: me.data,
			}),
		[
			people.data,
			journal.data,
			reminders.data,
			dates.data,
			gifts.data,
			audit.data,
			me.data,
		],
	);
	const isLoading = [people, journal, reminders, dates, gifts].some(
		(query) => query.isLoading,
	);
	const dashboardQueries = [people, journal, reminders, dates, gifts, audit];
	const hasError = dashboardQueries.some((query) => query.isError);
	const isStale = dashboardQueries.some((query) => query.isError && query.data);

	async function refresh(
		id:
			| DashboardSummaryCard["id"]
			| "pulse"
			| "actions"
			| "activity"
			| "moments",
	) {
		setRefreshingId(id);
		try {
			await invalidateDashboardQueries(queryClient, id);
		} finally {
			setRefreshingId(undefined);
		}
	}

	return (
		<div className="space-y-4">
			<header className="flex items-end justify-between">
				<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">
					Dashboard
				</h1>
				<span className="font-mono text-[11px] text-zinc-500">
					Last refreshed {formatTime(viewModel.lastUpdatedAt)}
				</span>
			</header>

			{hasError ? (
				<div className="rounded-md border border-amber-200 bg-amber-50 px-4 py-3 text-[13px] text-amber-800">
					Some dashboard data could not load. Available sections are shown with
					cached or empty data.
				</div>
			) : null}

			<SummaryCards
				cards={viewModel.summaryCards}
				isLoading={isLoading}
				isStale={isStale}
			/>

			<RelationshipPulseChart
				data={viewModel.pulse}
				isLoading={journal.isLoading}
				onRefresh={() => refresh("pulse")}
				isRefreshing={refreshingId === "pulse"}
			/>

			<div className="grid grid-cols-1 xl:grid-cols-2 gap-4">
				<ActionQueue
					actions={viewModel.actions}
					isLoading={reminders.isLoading || gifts.isLoading}
					onRefresh={() => refresh("actions")}
					isRefreshing={refreshingId === "actions"}
				/>
				<div className="space-y-4">
					<RecentRelationshipActivity
						activities={viewModel.activities}
						isLoading={journal.isLoading}
						onRefresh={() => refresh("activity")}
						isRefreshing={refreshingId === "activity"}
					/>
					<UpcomingMoments
						moments={viewModel.moments}
						isLoading={dates.isLoading}
						onRefresh={() => refresh("moments")}
						isRefreshing={refreshingId === "moments"}
					/>
				</div>
			</div>
		</div>
	);
}

async function invalidateDashboardQueries(
	queryClient: QueryClient,
	id: DashboardSummaryCard["id"] | "pulse" | "actions" | "activity" | "moments",
) {
	const map = {
		people: [keys.people.all],
		followups: [keys.reminders.all],
		dates: [keys.dates.all],
		gifts: [keys.gifts.all],
		journal: [keys.journal.all],
		pulse: [keys.journal.all],
		actions: [keys.reminders.all, keys.gifts.all],
		activity: [keys.journal.all],
		moments: [keys.dates.all],
	} as const;

	await Promise.all(
		map[id].map((queryKey) => queryClient.invalidateQueries({ queryKey })),
	);
}

function formatTime(value: string): string {
	return new Intl.DateTimeFormat(undefined, {
		hour: "numeric",
		minute: "2-digit",
	}).format(new Date(value));
}
