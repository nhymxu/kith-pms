import type { RelationshipGraph } from "../endpoints/relationships";

// d3-force mutates the arrays it receives (adds x/y/vx/vy, replaces link.source/target
// from number to node object). Passing cached query data directly corrupts the cache.
// Always feed a deep clone to ForceGraph2D.
export function cloneGraphData(g: RelationshipGraph): RelationshipGraph {
	return {
		nodes: g.nodes.map((n) => ({ ...n })),
		links: g.links.map((l) => ({ ...l })),
	};
}
