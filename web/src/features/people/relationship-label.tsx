import { useQuery } from "@tanstack/react-query";
import { Info } from "lucide-react";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { Badge } from "#/components/ui/badge";
import {
	Popover,
	PopoverContent,
	PopoverTrigger,
} from "#/components/ui/popover";
import { getSettings } from "#/endpoints/settings";
import { getUserPrefs, type RelationshipDisplay } from "#/lib/format-datetime";

// Shares the ["settings"] cache entry with the settings pages, so a saved
// change applies on next render; localStorage covers the first paint.
export function useRelationshipDisplay(): RelationshipDisplay {
	const { data } = useQuery({
		queryKey: ["settings"],
		queryFn: getSettings,
		staleTime: 5 * 60 * 1000,
	});
	return data?.relationship_display ?? getUserPrefs().relationshipDisplay;
}

function DirectionHint({ text }: { text: string }) {
	const { t } = useTranslation("people");
	return (
		<Popover>
			<PopoverTrigger
				openOnHover
				aria-label={t("relationship.showDirection")}
				className="text-sub hover:text-ink shrink-0 rounded-full focus-visible:outline-2 focus-visible:outline-ring"
			>
				<Info className="size-3.5" />
			</PopoverTrigger>
			<PopoverContent className="w-auto px-3 py-2 text-[12px] text-ink whitespace-pre-line">
				{text}
			</PopoverContent>
		</Popover>
	);
}

interface RelationshipLabelProps {
	display: RelationshipDisplay;
	// A row's type describes the other person relative to the profile owner:
	// "em ruột" on A's page means B is A's em ruột.
	typeName: string;
	inverseTypeName: string;
	selfName: string;
	otherName: string;
	other: ReactNode;
}

export function RelationshipLabel({
	display,
	typeName,
	inverseTypeName,
	selfName,
	otherName,
	other,
}: RelationshipLabelProps) {
	const { t } = useTranslation("people");
	const sentence = t("relationship.sentence", {
		type: typeName,
		self: selfName,
	});
	const tooltip = t("relationship.tooltip", {
		other: otherName,
		type: typeName,
		self: selfName,
		inverse: inverseTypeName || typeName,
	});

	switch (display) {
		case "sentence":
			return (
				<span className="flex-1 min-w-0">
					{other} <span className="text-sub">{sentence}</span>
				</span>
			);
		case "badge":
			return (
				<>
					<span className="flex-1 min-w-0">{other}</span>
					<span className="flex items-center gap-1 shrink-0">
						<DirectionHint text={tooltip} />
						<Badge variant="neutral">
							<span aria-hidden="true">→</span>
							{typeName}
						</Badge>
					</span>
				</>
			);
		default:
			return (
				<>
					<span className="flex items-center gap-1 shrink-0">
						<DirectionHint text={tooltip} />
						<Badge variant="neutral">{typeName}</Badge>
					</span>
					<span className="flex-1 min-w-0">{other}</span>
				</>
			);
	}
}
