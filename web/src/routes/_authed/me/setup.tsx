import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { Alert, AlertDescription } from "#/components/ui/alert";
import { Button } from "#/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "#/components/ui/card";
import { Input } from "#/components/ui/input";
import { setupMe } from "#/endpoints/me";
import { listPeople } from "#/endpoints/people";
import { keys } from "#/query-keys";

export const Route = createFileRoute("/_authed/me/setup")({
	component: MeSetupPage,
});

function MeSetupPage() {
	const navigate = useNavigate();
	const qc = useQueryClient();
	const [q, setQ] = useState("");
	const [selected, setSelected] = useState<number | null>(null);
	const [apiError, setApiError] = useState<string | null>(null);

	const { data: peopleList } = useQuery({
		queryKey: keys.people.list({ q: q || undefined }),
		queryFn: () => listPeople({ q: q || undefined, page_size: 30 }),
	});

	const mutation = useMutation({
		mutationFn: (personId: number) => setupMe(personId),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.me.all });
			navigate({ to: "/me" });
		},
		onError: (e) =>
			setApiError(e instanceof Error ? e.message : "Setup failed"),
	});

	const people = peopleList?.items ?? [];
	const selectedPerson = people.find((p) => p.id === selected);

	return (
		<div className="space-y-4 max-w-[480px]">
			<h1 className="text-[24px] font-semibold tracking-tight text-zinc-900">
				Who are you?
			</h1>
			<p className="text-[13px] text-zinc-500">
				Select which person in your contacts represents you.
			</p>

			{apiError && (
				<Alert variant="destructive">
					<AlertDescription>{apiError}</AlertDescription>
				</Alert>
			)}

			<Input
				placeholder="Search people…"
				value={q}
				onChange={(e) => setQ(e.target.value)}
			/>

			<div className="space-y-2 max-h-80 overflow-y-auto">
				{people.map((p) => (
					<button
						key={p.id}
						type="button"
						onClick={() => setSelected(p.id)}
						className={[
							"w-full text-left border rounded-md px-3 py-2.5 transition-colors text-[13px]",
							selected === p.id
								? "border-indigo-600 bg-indigo-50 text-zinc-900"
								: "border-zinc-200 hover:border-zinc-300 text-zinc-700",
						].join(" ")}
					>
						<p className="font-medium">{p.name}</p>
						{p.nickname && (
							<p className="text-[11px] text-zinc-400">"{p.nickname}"</p>
						)}
					</button>
				))}
				{people.length === 0 && (
					<p className="text-[13px] text-zinc-500">No people found.</p>
				)}
			</div>

			{selectedPerson && (
				<Card>
					<CardHeader className="pb-2">
						<CardTitle className="text-[13px] font-medium text-zinc-900">
							Selected: {selectedPerson.name}
						</CardTitle>
					</CardHeader>
					<CardContent>
						<Button
							onClick={() => mutation.mutate(selectedPerson.id)}
							disabled={mutation.isPending}
						>
							{mutation.isPending ? "Setting up…" : "Confirm selection"}
						</Button>
					</CardContent>
				</Card>
			)}
		</div>
	);
}
