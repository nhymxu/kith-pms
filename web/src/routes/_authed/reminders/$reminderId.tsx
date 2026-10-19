import {
	useMutation,
	useQueryClient,
	useSuspenseQuery,
} from "@tanstack/react-query";
import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Badge } from "#/components/ui/badge";
import { Button } from "#/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "#/components/ui/card";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import { deleteReminder, getReminder } from "#/endpoints/reminders";
import { CompleteButton } from "#/features/reminders/complete-button";
import { formatDate, formatDateTime } from "#/lib/format-datetime";
import { keys } from "#/query-keys";

export const Route = createFileRoute("/_authed/reminders/$reminderId")({
	component: ReminderDetailPage,
	pendingComponent: ReminderPending,
	errorComponent: ReminderError,
});

function ReminderPending() {
	const { t } = useTranslation("common");
	return <p className="text-sm font-base text-foreground/60">{t("loading")}</p>;
}

function ReminderError() {
	const { t } = useTranslation("reminders");
	return (
		<p className="text-sm font-base text-destructive">{t("page.notFound")}</p>
	);
}

function ReminderDetailPage() {
	const { t } = useTranslation("reminders");
	const { reminderId } = Route.useParams();
	const id = Number(reminderId);
	const navigate = useNavigate();
	const qc = useQueryClient();
	const [confirmOpen, setConfirmOpen] = useState(false);

	const { data } = useSuspenseQuery({
		queryKey: keys.reminders.detail(id),
		queryFn: () => getReminder(id),
	});

	const deleteMutation = useMutation({
		mutationFn: () => deleteReminder(id),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.reminders.all });
			navigate({ to: "/reminders" });
		},
	});

	const isOverdue = !data.completed && new Date(data.due_date) < new Date();

	return (
		<div className="max-w-lg space-y-4">
			<div className="flex items-center justify-between">
				<h1 className="text-[18px] font-semibold tracking-tight text-ink font-display">
					{data.title}
				</h1>
				<div className="flex gap-2">
					{!data.completed && <CompleteButton reminderId={id} />}
					<Button variant="neutral" asChild>
						<Link to="/reminders/$reminderId/edit" params={{ reminderId }}>
							{t("detail.edit")}
						</Link>
					</Button>
					<Button variant="destructive" onClick={() => setConfirmOpen(true)}>
						{t("delete.confirm")}
					</Button>
				</div>
			</div>

			<Card>
				<CardHeader>
					<CardTitle className="text-base">{t("detail.details")}</CardTitle>
				</CardHeader>
				<CardContent className="space-y-2 text-sm font-base">
					<div className="flex gap-2">
						<span className="text-foreground/60 w-28 shrink-0">
							{t("detail.status")}
						</span>
						{data.completed ? (
							<Badge variant="neutral">{t("status.done")}</Badge>
						) : isOverdue ? (
							<Badge variant="destructive">{t("status.overdue")}</Badge>
						) : (
							<Badge>{t("status.upcoming")}</Badge>
						)}
					</div>
					<div className="flex gap-2">
						<span className="text-foreground/60 w-28 shrink-0">
							{t("detail.due")}
						</span>
						<span>
							{data.due_date &&
							data.due_date.length > 10 &&
							data.due_date.slice(11, 16) !== "00:00"
								? formatDateTime(data.due_date)
								: formatDate(data.due_date)}
						</span>
					</div>
					{data.person_name && (
						<div className="flex gap-2">
							<span className="text-foreground/60 w-28 shrink-0">
								{t("detail.person")}
							</span>
							<span>{data.person_name}</span>
						</div>
					)}
					{data.notes && (
						<div className="flex gap-2">
							<span className="text-foreground/60 w-28 shrink-0">
								{t("detail.notes")}
							</span>
							<span className="whitespace-pre-wrap">{data.notes}</span>
						</div>
					)}
				</CardContent>
			</Card>

			<Dialog open={confirmOpen} onOpenChange={setConfirmOpen}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>{t("delete.title")}</DialogTitle>
						<DialogDescription>
							{t("delete.description", { title: data.title })}
						</DialogDescription>
					</DialogHeader>
					<DialogFooter>
						<Button variant="neutral" onClick={() => setConfirmOpen(false)}>
							{t("delete.cancel")}
						</Button>
						<Button
							variant="destructive"
							onClick={() => deleteMutation.mutate()}
							disabled={deleteMutation.isPending}
						>
							{deleteMutation.isPending
								? t("delete.deleting")
								: t("delete.confirm")}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</div>
	);
}
