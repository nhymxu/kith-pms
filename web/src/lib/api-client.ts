// Typed fetch boundary for all /v1 API calls.
// Never use raw fetch outside this module.

export class ApiError extends Error {
	readonly status: number
	readonly code: string

	constructor(status: number, code: string, message: string) {
		super(message)
		this.name = "ApiError"
		this.status = status
		this.code = code
	}
}

// ---- session-lost event bus ------------------------------------------------

type SessionLostHandler = () => void
const sessionLostHandlers = new Set<SessionLostHandler>()

export function onSessionLost(handler: SessionLostHandler): () => void {
	sessionLostHandlers.add(handler)
	return () => sessionLostHandlers.delete(handler)
}

function fireSessionLost() {
	for (const h of sessionLostHandlers) h()
}

// ---- core fetch helper -----------------------------------------------------

const BASE_URL = (import.meta.env.VITE_API_BASE_URL as string | undefined) ?? ""

export async function apiFetch<T = unknown>(
	path: string,
	init: RequestInit & { skipSessionLost?: boolean } = {},
): Promise<T> {
	const { skipSessionLost, ...fetchInit } = init
	init = fetchInit
	const method = (init.method ?? "GET").toUpperCase()
	const isReadMethod = method === "GET" || method === "HEAD"

	const headers = new Headers(init.headers)
	headers.set("Accept", "application/json")

	if (!isReadMethod) {
		headers.set("X-Requested-With", "kith-spa")
		// Skip Content-Type for FormData — browser sets it with boundary automatically.
		if (!(init.body instanceof FormData)) {
			headers.set("Content-Type", "application/json")
		}
	}

	const response = await fetch(`${BASE_URL}${path}`, {
		...init,
		credentials: "include",
		headers,
	})

	if (response.status === 204) {
		return undefined as T
	}

	if (!response.ok) {
		let code = String(response.status)
		let message = response.statusText || "Request failed"

		try {
			const body = (await response.json()) as { error?: string }
			if (body.error) {
				message = body.error
				code = body.error
			}
		} catch {
			// ignore parse error; use defaults
		}

		const err = new ApiError(response.status, code, message)

		if (response.status === 401 && !skipSessionLost) {
			fireSessionLost()
		}

		throw err
	}

	return response.json() as Promise<T>
}
