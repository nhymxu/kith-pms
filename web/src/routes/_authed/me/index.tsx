import { useSuspenseQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";
import { Badge } from "#/components/ui/badge";
import { Button } from "#/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "#/components/ui/card";
import { getMe } from "#/endpoints/me";
import { getAvatarUrl } from "#/endpoints/people";
import { keys } from "#/query-keys";

// 404 from getMe means Me is not set up yet — a valid state, not a load failure,
// so the scoped errorComponent reproduces the "set up" CTA rather than a generic error.
function MeNotSetUp() {
	const { t } = useTranslation("me");
	return (
		<div className="space-y-4 max-w-md">
			<h1 className="text-[18px] font-semibold tracking-tight text-ink font-display">
				{t("title")}
			</h1>
			<Card>
				<CardContent className="pt-6 space-y-3">
					<p className="text-sm font-base">{t("notSetUp")}</p>
					<Button asChild>
						<Link to="/me/setup">{t("setUpAction")}</Link>
					</Button>
				</CardContent>
			</Card>
		</div>
	);
}

function MePending() {
	const { t } = useTranslation("me");
	return <p className="text-[13px] text-sub">{t("loading")}</p>;
}

export const Route = createFileRoute("/_authed/me/")({
	component: MePage,
	pendingComponent: MePending,
	errorComponent: MeNotSetUp,
});

function MePage() {
	const { t } = useTranslation("me");
	const { data } = useSuspenseQuery({
		queryKey: keys.me.profile(),
		queryFn: getMe,
		retry: false,
	});

	return (
		<div className="space-y-4 max-w-md">
			<h1 className="text-[18px] font-semibold tracking-tight text-ink font-display">
				{t("title")}
			</h1>
			<Card>
				<CardHeader className="flex flex-row items-center gap-4">
					<div className="size-16 rounded-full overflow-hidden bg-chip flex items-center justify-center text-2xl font-medium text-chip-fg shrink-0">
						{data.avatar_path ? (
							<img
								src={getAvatarUrl(data.id)}
								alt={data.name}
								className="size-full object-cover"
							/>
						) : (
							data.name.charAt(0).toUpperCase()
						)}
					</div>
					<div>
						<CardTitle>{data.name}</CardTitle>
						{data.nickname && (
							<p className="text-[12px] text-sub">"{data.nickname}"</p>
						)}
					</div>
				</CardHeader>
				<CardContent className="space-y-3 text-[13px]">
					{data.date_of_birth && (
						<div className="flex gap-2">
							<span className="text-sub w-32 shrink-0">{t("dateOfBirth")}</span>
							<span className="font-mono text-ink">{data.date_of_birth}</span>
						</div>
					)}
					{data.labels && data.labels.length > 0 && (
						<div className="flex flex-wrap gap-1 pt-1">
							{data.labels.map((l) => (
								<Badge
									key={l.id}
									variant="neutral"
									style={{ borderColor: l.color }}
								>
									{l.name}
								</Badge>
							))}
						</div>
					)}
					<div className="pt-2 flex gap-2">
						<Button variant="neutral" size="sm" asChild>
							<Link
								to="/people/$personId/edit"
								params={{ personId: String(data.id) }}
							>
								{t("editProfile")}
							</Link>
						</Button>
						<Button variant="neutral" size="sm" asChild>
							<Link to="/me/setup">{t("changeMe")}</Link>
						</Button>
					</div>
				</CardContent>
			</Card>
		</div>
	);
}
