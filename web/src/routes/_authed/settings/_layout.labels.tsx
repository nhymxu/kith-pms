import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/_authed/settings/_layout/labels")({
	beforeLoad: () => {
		throw redirect({ to: "/settings/people-labels" });
	},
	component: () => null,
});
