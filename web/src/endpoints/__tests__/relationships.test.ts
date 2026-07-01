import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
	bulkCreateRelationships,
	getRelationshipGraph,
} from "../relationships";

const GRAPH_PAYLOAD = {
	nodes: [{ id: 1, name: "Alice", avatar: "", group: "", is_self: false }],
	links: [],
};

function mockFetch(body: unknown, status = 200) {
	return vi.spyOn(globalThis, "fetch").mockResolvedValue({
		status,
		ok: status >= 200 && status < 300,
		json: () => Promise.resolve({ data: body }),
		headers: new Headers({ "content-type": "application/json" }),
	} as unknown as Response);
}

describe("getRelationshipGraph", () => {
	let fetchSpy: ReturnType<typeof vi.spyOn>;

	beforeEach(() => {
		fetchSpy = mockFetch(GRAPH_PAYLOAD);
	});

	afterEach(() => {
		vi.restoreAllMocks();
	});

	it("calls /v1/relationships/graph without person_id", async () => {
		await getRelationshipGraph();
		const url = (fetchSpy.mock.calls[0]?.[0] as string) ?? "";
		expect(url).toContain("/v1/relationships/graph");
		expect(url).not.toContain("person_id");
	});

	it("appends ?person_id= when provided", async () => {
		await getRelationshipGraph(7);
		const url = (fetchSpy.mock.calls[0]?.[0] as string) ?? "";
		expect(url).toContain("/v1/relationships/graph?person_id=7");
	});

	it("unwraps the data envelope", async () => {
		const result = await getRelationshipGraph();
		expect(result.nodes).toEqual(GRAPH_PAYLOAD.nodes);
		expect(result.links).toEqual(GRAPH_PAYLOAD.links);
	});
});

function makePairs(n: number) {
	return Array.from({ length: n }, (_, i) => ({
		to_person_id: i + 1,
		relationship_type_id: 1,
	}));
}

describe("bulkCreateRelationships chunking", () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	it("sends 1 request for ≤50 pairs", async () => {
		const spy = vi.spyOn(globalThis, "fetch").mockResolvedValue({
			status: 200,
			ok: true,
			json: () => Promise.resolve({ data: { created: 10, skipped: 0 } }),
			headers: new Headers({ "content-type": "application/json" }),
		} as unknown as Response);

		await bulkCreateRelationships(1, makePairs(10));
		expect(spy).toHaveBeenCalledTimes(1);
	});

	it("sends 3 sequential requests for 120 pairs (50/50/20)", async () => {
		const spy = vi.spyOn(globalThis, "fetch").mockResolvedValue({
			status: 200,
			ok: true,
			json: () => Promise.resolve({ data: { created: 50, skipped: 0 } }),
			headers: new Headers({ "content-type": "application/json" }),
		} as unknown as Response);

		await bulkCreateRelationships(1, makePairs(120));
		expect(spy).toHaveBeenCalledTimes(3);

		const bodies = spy.mock.calls.map((call) => {
			const init = call[1] as RequestInit;
			return JSON.parse(init.body as string) as {
				relationships: unknown[];
			};
		});
		expect(bodies[0].relationships).toHaveLength(50);
		expect(bodies[1].relationships).toHaveLength(50);
		expect(bodies[2].relationships).toHaveLength(20);
	});

	it("sums created and skipped across chunks", async () => {
		vi.spyOn(globalThis, "fetch").mockResolvedValue({
			status: 200,
			ok: true,
			json: () => Promise.resolve({ data: { created: 30, skipped: 5 } }),
			headers: new Headers({ "content-type": "application/json" }),
		} as unknown as Response);

		const result = await bulkCreateRelationships(1, makePairs(60));
		expect(result.created).toBe(60);
		expect(result.skipped).toBe(10);
	});

	it("attaches partial totals to error when a chunk fails", async () => {
		let call = 0;
		vi.spyOn(globalThis, "fetch").mockImplementation(async () => {
			call++;
			if (call === 2) {
				return {
					status: 500,
					ok: false,
					json: () => Promise.resolve({ error: "server error" }),
					headers: new Headers({ "content-type": "application/json" }),
				} as unknown as Response;
			}
			return {
				status: 200,
				ok: true,
				json: () => Promise.resolve({ data: { created: 50, skipped: 0 } }),
				headers: new Headers({ "content-type": "application/json" }),
			} as unknown as Response;
		});

		const err = await bulkCreateRelationships(1, makePairs(120)).catch(
			(e) => e,
		);
		expect(err).toBeDefined();
		expect((err as { partial?: { created: number } }).partial?.created).toBe(
			50,
		);
	});
});
