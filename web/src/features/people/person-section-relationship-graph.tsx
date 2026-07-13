import { useSuspenseQuery } from "@tanstack/react-query";
import { lazy, Suspense } from "react";
import { getRelationshipGraph } from "#/endpoints/relationships";
import { keys } from "#/query-keys";

const LazyRelationshipGraph = lazy(
	() => import("#/features/network/relationship-graph"),
);

function SectionHeading({ children }: { children: React.ReactNode }) {
	return (
		<h2 className="text-[11px] font-semibold uppercase tracking-widest text-sub mb-2">
			{children}
		</h2>
	);
}

interface EgoGraphInnerProps {
	personId: number;
}

function EgoGraphInner({ personId }: EgoGraphInnerProps) {
	const { data } = useSuspenseQuery({
		queryKey: keys.relationships.graph(personId),
		queryFn: () => getRelationshipGraph(personId),
	});

	if (data.nodes.length <= 1) {
		return <p className="text-sm text-sub">No connections yet.</p>;
	}

	return (
		<LazyRelationshipGraph data={data} focusNodeId={personId} height={360} />
	);
}

interface PersonSectionRelationshipGraphProps {
	personId: number;
}

export function PersonSectionRelationshipGraph({
	personId,
}: PersonSectionRelationshipGraphProps) {
	return (
		<div>
			<SectionHeading>Relationship Network</SectionHeading>
			<Suspense
				fallback={
					<div className="flex items-center justify-center rounded-md border-bw border-line bg-muted py-8 text-[13px] text-sub">
						Loading…
					</div>
				}
			>
				<EgoGraphInner personId={personId} />
			</Suspense>
		</div>
	);
}
