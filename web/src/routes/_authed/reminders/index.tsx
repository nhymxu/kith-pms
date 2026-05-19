import { useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";
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
	const [tab, setTab] = useState<ReminderStatus>("upcoming");
	const qc = useQueryClient();

	const { data, isPending, isError } = useQuery({
		queryKey: [
			...keys.reminders.list({
				completed:
					tab === "all" ? undefined : tab === "upcoming" ? false : undefined,
			}),
			tab,
		],
		queryFn: () => listReminders({ status: tab }),
	});

	if (isError)
		return (
			<p className="text-[13px] text-red-600">Failed to load reminders.</p>
		);

	return (
		<div className="space-y-4">
			<div className="flex items-center justify-between">
				<h1 className="text-[18px] font-semibold tracking-tight text-zinc-900">
					Reminders
				</h1>
				<Button asChild>
					<Link to="/reminders/new">New Reminder</Link>
				</Button>
			</div>

			<Tabs value={tab} onValueChange={(v) => setTab(v as ReminderStatus)}>
				<TabsList>
					<TabsTrigger value="upcoming">Upcoming</TabsTrigger>
					<TabsTrigger value="overdue">Overdue</TabsTrigger>
					<TabsTrigger value="all">All</TabsTrigger>
				</TabsList>

				<TabsContent value={tab}>
					{isPending ? (
						<p className="text-[13px] text-zinc-500 py-4">Loading…</p>
					) : (
						<RemindersTable
							data={data ?? []}
							onCompleted={() =>
								qc.invalidateQueries({ queryKey: keys.reminders.all })
							}
						/>
					)}
				</TabsContent>
			</Tabs>
		</div>
	);
}
