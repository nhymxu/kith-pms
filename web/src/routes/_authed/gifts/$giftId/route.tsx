import { createFileRoute, Outlet } from "@tanstack/react-router"

export const Route = createFileRoute("/_authed/gifts/$giftId")({
	component: () => <Outlet />,
})
