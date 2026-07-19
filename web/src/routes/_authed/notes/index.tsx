import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { Button } from "#/components/ui/button";
import { Card, CardContent } from "#/components/ui/card";
import { getMe } from "#/endpoints/me";
import { NotesList } from "#/features/notes/notes-list";
import { keys } from "#/query-keys";

// 404 from getMe means Me is not set up yet — a valid state, not a load failure,
// so the scoped errorComponent reproduces the "set up" CTA rather than a generic error.
function NotesNotSetUp() {
	return (
		<div className="space-y-4 max-w-2xl">
			<h1 className="text-[18px] font-semibold tracking-tight text-ink font-display">
				Notes
			</h1>
			<Card>
				<CardContent className="pt-6 space-y-3">
					<p className="text-sm font-base">
						You haven't set up your self-profile yet. Pick an existing person to
						represent yourself before adding self notes.
					</p>
					<Button asChild>
						<Link to="/me/setup">Set up my profile</Link>
					</Button>
				</CardContent>
			</Card>
		</div>
	);
}

export const Route = createFileRoute("/_authed/notes/")({
	component: NotesPage,
	pendingComponent: () => <p className="text-[13px] text-sub">Loading…</p>,
	errorComponent: NotesNotSetUp,
});

function NotesPage() {
	const { data: self } = useSuspenseQuery({
		queryKey: keys.me.profile(),
		queryFn: getMe,
		retry: false,
	});

	return (
		<div className="space-y-4 max-w-2xl">
			<h1 className="text-[18px] font-semibold tracking-tight text-ink font-display">
				Notes
			</h1>
			<NotesList personId={self.id} />
		</div>
	);
}
