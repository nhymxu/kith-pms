import {
	keepPreviousData,
	useQuery,
	useQueryClient,
} from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "#/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "#/components/ui/tabs";
import type { ReminderStatus } from "#/endpoints/reminders";
import { listReminders } from "#/endpoints/reminders";
import { RemindersTable } from "#/features/reminders/reminders-table";
import { keys } from "#/query-keys";

export const Route = createFileRoute("/_authed/reminders/")({
	component: RemindersPage,
});

function RemindersPage() {
	const { t } = useTranslation("reminders");
	const [tab, setTab] = useState<ReminderStatus>("upcoming");
	const qc = useQueryClient();

	const { data } = useQuery({
		queryKey: [
			...keys.reminders.list({
				completed:
					tab === "all" ? undefined : tab === "upcoming" ? false : undefined,
			}),
			tab,
		],
		queryFn: () => listReminders({ status: tab }),
		placeholderData: keepPreviousData,
	});

	return (
		<div className="space-y-4">
			<div className="flex items-center justify-between">
				<h1 className="text-[18px] font-semibold tracking-tight text-ink font-display">
					{t("page.title")}
				</h1>
				<Button asChild>
					<Link to="/reminders/new">{t("page.newReminder")}</Link>
				</Button>
			</div>

			<Tabs value={tab} onValueChange={(v) => setTab(v as ReminderStatus)}>
				<TabsList>
					<TabsTrigger value="upcoming">{t("status.upcoming")}</TabsTrigger>
					<TabsTrigger value="overdue">{t("status.overdue")}</TabsTrigger>
					<TabsTrigger value="all">{t("page.tabAll")}</TabsTrigger>
				</TabsList>

				<TabsContent value={tab}>
					<RemindersTable
						data={data ?? []}
						onCompleted={() =>
							qc.invalidateQueries({ queryKey: keys.reminders.all })
						}
					/>
				</TabsContent>
			</Tabs>
		</div>
	);
}
