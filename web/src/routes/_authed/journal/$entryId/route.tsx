import { createFileRoute, Outlet } from "@tanstack/react-router";

export const Route = createFileRoute("/_authed/journal/$entryId")({
	component: () => <Outlet />,
});
