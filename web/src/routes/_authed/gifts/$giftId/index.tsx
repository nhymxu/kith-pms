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
import { deleteGift, getGift } from "#/endpoints/gifts";
import { formatCurrency } from "#/lib/format-currency";
import { formatDate } from "#/lib/format-datetime";
import { keys } from "#/query-keys";

export const Route = createFileRoute("/_authed/gifts/$giftId/")({
	component: GiftDetailPage,
	pendingComponent: GiftPending,
	errorComponent: GiftError,
});

function GiftPending() {
	const { t } = useTranslation("common");
	return <p className="text-sm font-base text-foreground/60">{t("loading")}</p>;
}

function GiftError() {
	const { t } = useTranslation("gifts");
	return (
		<p className="text-sm font-base text-destructive">{t("page.notFound")}</p>
	);
}

function GiftDetailPage() {
	const { t } = useTranslation("gifts");
	const { giftId } = Route.useParams();
	const id = Number(giftId);
	const navigate = useNavigate();
	const qc = useQueryClient();
	const [confirmOpen, setConfirmOpen] = useState(false);

	const { data } = useSuspenseQuery({
		queryKey: keys.gifts.detail(id),
		queryFn: () => getGift(id),
	});

	const deleteMutation = useMutation({
		mutationFn: () => deleteGift(id),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.gifts.all });
			navigate({ to: "/gifts" });
		},
	});

	const debtLabel =
		data.debt_type === "i_owe"
			? t("debt.iOwe")
			: data.debt_type === "they_owe"
				? t("debt.theyOwe")
				: "—";

	const amountLabel = formatCurrency(data.amount_cents, data.currency);

	return (
		<div className="max-w-lg space-y-4">
			<div className="flex items-center justify-between">
				<h1 className="text-[18px] font-semibold tracking-tight text-ink font-display">
					{data.title}
				</h1>
				<div className="flex gap-2">
					<Button variant="neutral" asChild>
						<Link to="/gifts/$giftId/edit" params={{ giftId }}>
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
							{t("detail.person")}
						</span>
						<Link
							to="/people/$personId"
							params={{ personId: String(data.person_id) }}
							className="text-accent-text hover:underline"
						>
							{data.person_name}
						</Link>
					</div>
					<div className="flex gap-2">
						<span className="text-foreground/60 w-28 shrink-0">
							{t("detail.direction")}
						</span>
						<Badge variant="neutral">
							{t(`direction.${data.direction}`, {
								defaultValue: data.direction,
							})}
						</Badge>
					</div>
					<div className="flex gap-2">
						<span className="text-foreground/60 w-28 shrink-0">
							{t("detail.debt")}
						</span>
						<span>{debtLabel}</span>
					</div>
					<div className="flex gap-2">
						<span className="text-foreground/60 w-28 shrink-0">
							{t("detail.date")}
						</span>
						<span>{formatDate(data.date)}</span>
					</div>
					<div className="flex gap-2">
						<span className="text-foreground/60 w-28 shrink-0">
							{t("detail.amount")}
						</span>
						<span>{amountLabel}</span>
					</div>
					<div className="flex gap-2">
						<span className="text-foreground/60 w-28 shrink-0">
							{t("detail.notes")}
						</span>
						<span className="whitespace-pre-wrap">{data.notes || "—"}</span>
					</div>
					<div className="flex gap-2">
						<span className="text-foreground/60 w-28 shrink-0">
							{t("detail.image")}
						</span>
						{data.image_path ? (
							<img
								src={`/v1/gifts/${data.id}/image`}
								alt={data.title}
								className="mt-1 rounded-base border-bw border-line max-h-64 object-contain"
							/>
						) : (
							<span className="text-sub">{t("detail.noImage")}</span>
						)}
					</div>
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
