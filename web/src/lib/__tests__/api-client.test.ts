// Unit tests for apiFetch: header injection, FormData branch, 401 session-lost signal.
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest"
import { apiFetch, ApiError, onSessionLost } from "../api-client"

// Helper to build a minimal Response-like object
function mockResponse(status: number, body: unknown = {}, ok?: boolean): Response {
	return {
		status,
		statusText: status === 200 ? "OK" : "Error",
		ok: ok ?? (status >= 200 && status < 300),
		json: () => Promise.resolve(body),
		headers: new Headers(),
	} as unknown as Response
}

describe("apiFetch", () => {
	let fetchSpy: ReturnType<typeof vi.spyOn>

	beforeEach(() => {
		fetchSpy = vi.spyOn(globalThis, "fetch")
	})

	afterEach(() => {
		vi.restoreAllMocks()
	})

	it("sets Accept and credentials on GET", async () => {
		fetchSpy.mockResolvedValueOnce(mockResponse(200, { data: "ok" }))

		await apiFetch("/v1/test")

		const [, init] = fetchSpy.mock.calls[0] as [string, RequestInit & { headers: Headers }]
		const headers = init.headers as Headers
		expect(headers.get("Accept")).toBe("application/json")
		expect(init.credentials).toBe("include")
	})

	it("injects X-Requested-With and Content-Type on non-GET JSON body", async () => {
		fetchSpy.mockResolvedValueOnce(mockResponse(200, { data: "ok" }))

		await apiFetch("/v1/test", { method: "POST", body: JSON.stringify({ foo: 1 }) })

		const [, init] = fetchSpy.mock.calls[0] as [string, RequestInit & { headers: Headers }]
		const headers = init.headers as Headers
		expect(headers.get("X-Requested-With")).toBe("kith-spa")
		expect(headers.get("Content-Type")).toBe("application/json")
	})

	it("skips Content-Type when body is FormData", async () => {
		fetchSpy.mockResolvedValueOnce(mockResponse(200, { data: "ok" }))

		const form = new FormData()
		form.append("file", new Blob(["x"]), "x.png")
		await apiFetch("/v1/test", { method: "POST", body: form })

		const [, init] = fetchSpy.mock.calls[0] as [string, RequestInit & { headers: Headers }]
		const headers = init.headers as Headers
		expect(headers.get("X-Requested-With")).toBe("kith-spa")
		expect(headers.get("Content-Type")).toBeNull()
	})

	it("does NOT inject X-Requested-With on GET", async () => {
		fetchSpy.mockResolvedValueOnce(mockResponse(200, { data: "ok" }))

		await apiFetch("/v1/test")

		const [, init] = fetchSpy.mock.calls[0] as [string, RequestInit & { headers: Headers }]
		const headers = init.headers as Headers
		expect(headers.get("X-Requested-With")).toBeNull()
	})

	it("throws ApiError for non-2xx responses", async () => {
		fetchSpy.mockResolvedValueOnce(mockResponse(422, { error: "name is required" }))

		await expect(apiFetch("/v1/test", { method: "POST", body: "{}" })).rejects.toThrow(ApiError)
	})

	it("fires onSessionLost listeners on 401", async () => {
		fetchSpy.mockResolvedValueOnce(mockResponse(401, { error: "unauthorized" }))

		const handler = vi.fn()
		const cleanup = onSessionLost(handler)

		await expect(apiFetch("/v1/test")).rejects.toThrow(ApiError)
		expect(handler).toHaveBeenCalledOnce()

		cleanup()
	})

	it("does NOT fire onSessionLost on 422", async () => {
		fetchSpy.mockResolvedValueOnce(mockResponse(422, { error: "invalid" }))

		const handler = vi.fn()
		const cleanup = onSessionLost(handler)

		await expect(apiFetch("/v1/test", { method: "POST", body: "{}" })).rejects.toThrow(ApiError)
		expect(handler).not.toHaveBeenCalled()

		cleanup()
	})

	it("returns undefined for 204 No Content", async () => {
		fetchSpy.mockResolvedValueOnce({ status: 204, ok: true } as unknown as Response)

		const result = await apiFetch("/v1/test", { method: "DELETE" })
		expect(result).toBeUndefined()
	})
})
