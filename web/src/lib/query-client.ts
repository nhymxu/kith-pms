import { QueryCache, QueryClient } from "@tanstack/react-query"
import { ApiError, onSessionLost } from "./api-client"

// Called by AuthProvider once it sets up the session-lost → clear pattern.
// Exported so auth-context can wire it after creating the client.
let _sessionLostCleanup: (() => void) | null = null

export function createQueryClient(): QueryClient {
	const queryClient = new QueryClient({
		defaultOptions: {
			queries: {
				staleTime: 30_000,
				refetchOnWindowFocus: false,
				retry: (failureCount, error) => {
					if (error instanceof ApiError && error.status >= 400 && error.status < 500) {
						return false
					}
					return failureCount < 1
				},
			},
		},
		queryCache: new QueryCache({
			onError: (error) => {
				if (error instanceof ApiError && error.status === 401) {
					// api-client already fired onSessionLost; nothing extra needed here.
					// Kept as a second safety net if a query throws directly.
				}
			},
		}),
	})

	// Wire session-lost bus: clear all cached queries when session expires.
	if (_sessionLostCleanup) _sessionLostCleanup()
	_sessionLostCleanup = onSessionLost(() => {
		queryClient.clear()
	})

	return queryClient
}
