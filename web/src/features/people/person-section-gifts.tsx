import { useSuspenseQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { Gift } from "lucide-react";
import { useState } from "react";
import { QueryBoundary } from "#/components/query-boundary";
import { Badge } from "#/components/ui/badge";
import { Button } from "#/components/ui/button";
import { listGifts } from "#/endpoints/gifts";
import { keys } from "#/query-keys";
import { QuickGiftDialog } from "./quick-actions";

function SectionHeading({ children }: { children: React.ReactNode }) {
	return (
		<h2 className="text-[11px] font-semibold uppercase tracking-widest text-zinc-400 mb-2">
			{children}
		</h2>
	);
}

interface GiftsListInnerProps {
	personId: number;
}

function GiftsListInner({ personId }: GiftsListInnerProps) {
	const { data } = useSuspenseQuery({
		queryKey: keys.gifts.list({ person_id: personId, page_size: 10 }),
		queryFn: () => listGifts({ person_id: personId, page_size: 10 }),
	});

	if (!data.items.length) {
		return <p className="text-sm text-zinc-400">No gifts.</p>;
	}

	return (
		<div className="space-y-2">
			{data.items.map((g) => (
				<Link
					key={g.id}
					to="/gifts/$giftId"
					params={{ giftId: String(g.id) }}
					className="flex items-center gap-3 text-sm border border-zinc-200 rounded-md p-2 hover:bg-zinc-50"
				>
					<span className="font-medium flex-1">{g.title}</span>
					<Badge variant="neutral">{g.direction}</Badge>
					{g.date && (
						<span className="font-mono text-[12px] text-zinc-500">
							{g.date}
						</span>
					)}
				</Link>
			))}
		</div>
	);
}

interface GiftsSectionProps {
	personId: number;
}

export function GiftsSection({ personId }: GiftsSectionProps) {
	const [giftOpen, setGiftOpen] = useState(false);

	return (
		<div>
			<div className="flex items-center justify-between mb-2">
				<SectionHeading>Gifts</SectionHeading>
				<Button variant="neutral" size="sm" onClick={() => setGiftOpen(true)}>
					<Gift className="size-3" /> Quick gift
				</Button>
			</div>
			<QueryBoundary>
				<GiftsListInner personId={personId} />
			</QueryBoundary>
			<QuickGiftDialog
				personId={personId}
				open={giftOpen}
				onClose={() => setGiftOpen(false)}
			/>
		</div>
	);
}
