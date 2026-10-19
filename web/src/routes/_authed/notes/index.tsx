import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "#/components/ui/button";
import { Card, CardContent } from "#/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "#/components/ui/tabs";
import { getMe } from "#/endpoints/me";
import { AllNotesList } from "#/features/notes/all-notes-list";
import { NotesList } from "#/features/notes/notes-list";
import { keys } from "#/query-keys";

export const Route = createFileRoute("/_authed/notes/")({
	component: NotesPage,
});

// 404 from getMe means Me is not set up yet: a valid state, so only the
// "My notes" tab shows the setup CTA and "All notes" stays usable.
function MyNotes() {
	const { t } = useTranslation("notes");
	const { data: self, isPending } = useQuery({
		queryKey: keys.me.profile(),
		queryFn: getMe,
		retry: false,
	});

	if (isPending)
		return <p className="text-[13px] text-sub">{t("page.loading")}</p>;

	if (!self) {
		return (
			<Card>
				<CardContent className="pt-6 space-y-3">
					<p className="text-sm font-base">{t("page.noSelfProfile")}</p>
					<Button asChild>
						<Link to="/me/setup">{t("page.setupProfile")}</Link>
					</Button>
				</CardContent>
			</Card>
		);
	}

	return <NotesList personId={self.id} paged />;
}

function NotesPage() {
	const { t } = useTranslation("notes");
	const [tab, setTab] = useState("mine");

	return (
		<div className="space-y-4 max-w-2xl">
			<h1 className="text-[18px] font-semibold tracking-tight text-ink font-display">
				{t("page.title")}
			</h1>
			<Tabs value={tab} onValueChange={setTab}>
				<TabsList>
					<TabsTrigger value="mine">{t("page.tabMine")}</TabsTrigger>
					<TabsTrigger value="all">{t("page.tabAll")}</TabsTrigger>
				</TabsList>
				<TabsContent value="mine">
					<MyNotes />
				</TabsContent>
				<TabsContent value="all">
					<AllNotesList />
				</TabsContent>
			</Tabs>
		</div>
	);
}
