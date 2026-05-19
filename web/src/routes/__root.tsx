import type { QueryClient } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { createRootRouteWithContext, Outlet } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import type { AuthState } from "../lib/auth-context-types";
import "../styles.css";

export const Route = createRootRouteWithContext<{
	queryClient: QueryClient;
	auth: AuthState;
}>()({
	component: RootComponent,
});

function RootComponent() {
	return (
		<>
			<Outlet />
			<TanStackRouterDevtools position="bottom-right" />
			<ReactQueryDevtools buttonPosition="bottom-left" />
		</>
	);
}
