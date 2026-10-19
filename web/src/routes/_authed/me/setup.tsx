import {
	useMutation,
	useQueryClient,
	useSuspenseQuery,
} from "@tanstack/react-query";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { QueryBoundary } from "#/components/query-boundary";
import { Alert, AlertDescription } from "#/components/ui/alert";
import { Button } from "#/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "#/components/ui/card";
import { Input } from "#/components/ui/input";
import { setupMe } from "#/endpoints/me";
import { listPeople } from "#/endpoints/people";
import { apiErrorMessage } from "#/lib/api-error-message";
import { keys } from "#/query-keys";

export const Route = createFileRoute("/_authed/me/setup")({
	component: MeSetupPage,
});

function MeSetupPage() {
	const { t } = useTranslation("me");
	const navigate = useNavigate();
	const qc = useQueryClient();
	const [q, setQ] = useState("");
	const [selected, setSelected] = useState<number | null>(null);
	const [apiError, setApiError] = useState<string | null>(null);

	const mutation = useMutation({
		mutationFn: (personId: number) => setupMe(personId),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.me.all });
			navigate({ to: "/me" });
		},
		onError: (e) => setApiError(apiErrorMessage(e, t("setup.failed"))),
	});

	return (
		<div className="space-y-4 max-w-[480px]">
			<h1 className="text-[24px] font-semibold tracking-tight text-ink font-display">
				{t("setup.title")}
			</h1>
			<p className="text-[13px] text-sub">{t("setup.subtitle")}</p>

			{apiError && (
				<Alert variant="destructive">
					<AlertDescription>{apiError}</AlertDescription>
				</Alert>
			)}

			<Input
				placeholder={t("setup.searchPlaceholder")}
				value={q}
				onChange={(e) => setQ(e.target.value)}
			/>

			<QueryBoundary>
				<PeopleSelectionList
					q={q}
					selected={selected}
					onSelect={setSelected}
					mutation={mutation}
				/>
			</QueryBoundary>
		</div>
	);
}

interface PeopleSelectionListProps {
	q: string;
	selected: number | null;
	onSelect: (id: number) => void;
	mutation: ReturnType<typeof useMutation<number, Error, number>>;
}

function PeopleSelectionList({
	q,
	selected,
	onSelect,
	mutation,
}: PeopleSelectionListProps) {
	const { t } = useTranslation("me");
	const { data: peopleList } = useSuspenseQuery({
		queryKey: keys.people.list({ q: q || undefined }),
		queryFn: () => listPeople({ q: q || undefined, page_size: 30 }),
	});

	const people = peopleList.items;
	const selectedPerson = people.find((p) => p.id === selected);

	return (
		<>
			<div className="space-y-2 max-h-80 overflow-y-auto">
				{people.map((p) => (
					<button
						key={p.id}
						type="button"
						onClick={() => onSelect(p.id)}
						className={[
							"w-full text-left border-bw rounded-base px-3 py-2.5 transition-colors text-[13px]",
							selected === p.id
								? "border-accent bg-nav-active text-ink"
								: "border-line hover:border-line-soft text-ink",
						].join(" ")}
					>
						<p className="font-medium">{p.name}</p>
						{p.nickname && (
							<p className="text-[11px] text-sub">"{p.nickname}"</p>
						)}
					</button>
				))}
				{people.length === 0 && (
					<p className="text-[13px] text-sub">{t("setup.noPeople")}</p>
				)}
			</div>

			{selectedPerson && (
				<Card>
					<CardHeader className="pb-2">
						<CardTitle className="text-[13px] font-medium">
							{t("setup.selected", { name: selectedPerson.name })}
						</CardTitle>
					</CardHeader>
					<CardContent>
						<Button
							onClick={() => mutation.mutate(selectedPerson.id)}
							disabled={mutation.isPending}
						>
							{mutation.isPending ? t("setup.confirming") : t("setup.confirm")}
						</Button>
					</CardContent>
				</Card>
			)}
		</>
	);
}
