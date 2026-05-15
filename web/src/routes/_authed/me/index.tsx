import { createFileRoute, Link } from "@tanstack/react-router"
import { useQuery } from "@tanstack/react-query"
import { getMe } from "#/endpoints/me"
import { getAvatarUrl } from "#/endpoints/people"
import { keys } from "#/query-keys"
import { Button } from "#/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "#/components/ui/card"
import { Badge } from "#/components/ui/badge"

export const Route = createFileRoute("/_authed/me/")({
	component: MePage,
})

function MePage() {
	const { data, isPending, isError } = useQuery({
		queryKey: keys.me.profile(),
		queryFn: getMe,
		retry: false,
	})

	if (isPending) return <p className="text-sm font-base text-foreground/60">Loading…</p>

	// 404 from getMe means Me is not set up yet
	if (isError || !data) {
		return (
			<div className="space-y-4 max-w-md">
				<h1 className="text-2xl font-heading">My Profile</h1>
				<Card>
					<CardContent className="pt-6 space-y-3">
						<p className="text-sm font-base">
							You haven't set up your self-profile yet. Pick an existing person to represent yourself.
						</p>
						<Button asChild>
							<Link to="/me/setup">Set up my profile</Link>
						</Button>
					</CardContent>
				</Card>
			</div>
		)
	}

	return (
		<div className="space-y-4 max-w-md">
			<h1 className="text-2xl font-heading">My Profile</h1>
			<Card>
				<CardHeader className="flex flex-row items-center gap-4">
					<div className="size-16 rounded-base border-2 border-border overflow-hidden bg-secondary-background flex items-center justify-center text-2xl font-heading shrink-0">
						{data.avatar_path
							? <img src={getAvatarUrl(data.id)} alt={data.name} className="size-full object-cover" />
							: data.name.charAt(0).toUpperCase()
						}
					</div>
					<div>
						<CardTitle>{data.name}</CardTitle>
						{data.nickname && <p className="text-sm text-foreground/60">"{data.nickname}"</p>}
					</div>
				</CardHeader>
				<CardContent className="space-y-3 text-sm font-base">
					{data.relationship_type && (
						<div className="flex gap-2">
							<span className="text-foreground/60 w-32 shrink-0">Relationship</span>
							<span>{data.relationship_type}</span>
						</div>
					)}
					{data.date_of_birth && (
						<div className="flex gap-2">
							<span className="text-foreground/60 w-32 shrink-0">Date of birth</span>
							<span>{data.date_of_birth}</span>
						</div>
					)}
					{data.labels && data.labels.length > 0 && (
						<div className="flex flex-wrap gap-1 pt-1">
							{data.labels.map((l) => (
								<Badge key={l.id} variant="neutral" style={{ borderColor: l.color }}>{l.name}</Badge>
							))}
						</div>
					)}
					<div className="pt-2 flex gap-2">
						<Button variant="neutral" size="sm" asChild>
							<Link to="/people/$personId/edit" params={{ personId: String(data.id) }}>
								Edit profile
							</Link>
						</Button>
						<Button variant="neutral" size="sm" asChild>
							<Link to="/me/setup">Change Me person</Link>
						</Button>
					</div>
				</CardContent>
			</Card>
		</div>
	)
}
