import { describe, expect, it } from "vitest";
import type { RelationshipGraph } from "../../endpoints/relationships";
import { cloneGraphData } from "../graph-data";

const sample: RelationshipGraph = {
	nodes: [
		{
			id: 1,
			name: "Alice",
			nickname: "",
			avatar: "",
			group: "Friend",
			groups: ["Friend"],
			is_self: false,
		},
		{
			id: 2,
			name: "Bob",
			nickname: "bobby",
			avatar: "",
			group: "",
			groups: [],
			is_self: true,
		},
	],
	links: [{ source: 1, target: 2, type: "Friend", reverse_type: "Friend" }],
};

describe("cloneGraphData", () => {
	it("returns new arrays (not reference-equal)", () => {
		const clone = cloneGraphData(sample);
		expect(clone.nodes).not.toBe(sample.nodes);
		expect(clone.links).not.toBe(sample.links);
	});

	it("returns new node objects", () => {
		const clone = cloneGraphData(sample);
		expect(clone.nodes[0]).not.toBe(sample.nodes[0]);
	});

	it("returns new link objects", () => {
		const clone = cloneGraphData(sample);
		expect(clone.links[0]).not.toBe(sample.links[0]);
	});

	it("source stays a number (not mutated to object)", () => {
		const clone = cloneGraphData(sample);
		expect(typeof clone.links[0]?.source).toBe("number");
	});

	it("cloned values match original", () => {
		const clone = cloneGraphData(sample);
		expect(clone.nodes[0]?.name).toBe("Alice");
		expect(clone.links[0]?.type).toBe("Friend");
	});
});
