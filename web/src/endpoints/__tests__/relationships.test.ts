import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { getRelationshipGraph } from "../relationships";

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
