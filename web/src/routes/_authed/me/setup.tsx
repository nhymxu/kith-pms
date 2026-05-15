import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { useState } from "react"
import { listPeople } from "#/endpoints/people"
import { setupMe } from "#/endpoints/me"
import { keys } from "#/query-keys"
import { Button } from "#/components/ui/button"
import { Alert, AlertDescription } from "#/components/ui/alert"
import { Card, CardContent, CardHeader, CardTitle } from "#/components/ui/card"
import { Input } from "#/components/ui/input"

export const Route = createFileRoute("/_authed/me/setup")({
	component: MeSetupPage,
})

function MeSetupPage() {
	const navigate = useNavigate()
	const qc = useQueryClient()
	const [q, setQ] = useState("")
	const [selected, setSelected] = useState<number | null>(null)
	const [apiError, setApiError] = useState<string | null>(null)

	const { data: peopleList } = useQuery({
		queryKey: keys.people.list({ q: q || undefined }),
		queryFn: () => listPeople({ q: q || undefined, page_size: 30 }),
	})

	const mutation = useMutation({
		mutationFn: (personId: number) => setupMe(personId),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.me.all })
			navigate({ to: "/me" })
		},
		onError: (e) => setApiError(e instanceof Error ? e.message : "Setup failed"),
	})

	const people = peopleList?.items ?? []
	const selectedPerson = people.find((p) => p.id === selected)

	return (
		<div className="space-y-4 max-w-md">
			<h1 className="text-2xl font-heading">Set Up My Profile</h1>
			<p className="text-sm font-base text-foreground/60">
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
							"w-full text-left border-2 rounded-base p-3 transition-colors text-sm",
							selected === p.id
								? "border-main bg-main/10"
								: "border-border hover:border-main/50",
						].join(" ")}
					>
						<p className="font-heading">{p.name}</p>
						{p.nickname && <p className="text-xs text-foreground/50">"{p.nickname}"</p>}
					</button>
				))}
				{people.length === 0 && (
					<p className="text-sm text-foreground/50">No people found.</p>
				)}
			</div>

			{selectedPerson && (
				<Card>
					<CardHeader className="pb-2">
						<CardTitle className="text-sm font-heading">Selected: {selectedPerson.name}</CardTitle>
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
	)
}
