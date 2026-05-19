import type { QueryClient } from "@tanstack/react-query";
import { createRouter as createTanStackRouter } from "@tanstack/react-router";
import type { AuthState } from "./lib/auth-context-types";
import { routeTree } from "./routeTree.gen";

export interface RouterContext {
	queryClient: QueryClient;
	auth: AuthState;
}

// Safe initial auth placeholder — prevents TypeError in any route loader that
// runs before the first React commit (none exist today, but guards future work).
const authPlaceholder: AuthState = {
	user: null,
	isLoading: true,
	login: async () => {
		throw new Error("AuthProvider not yet mounted");
	},
	logout: async () => {},
	logoutAll: async () => {},
	refresh: async () => {},
};

// Router is created once with placeholder context values.
// Live values are injected via <RouterProvider context={...} />.
export function getRouter(queryClient: QueryClient) {
	return createTanStackRouter({
		routeTree,
		defaultPreload: "intent",
		defaultPreloadStaleTime: 0,
		scrollRestoration: true,
		context: { queryClient, auth: authPlaceholder },
	});
}

declare module "@tanstack/react-router" {
	interface Register {
		router: ReturnType<typeof getRouter>;
	}
}
