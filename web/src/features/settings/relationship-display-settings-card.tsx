import { useMutation, type useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { Button } from "#/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "#/components/ui/card";
import { Label } from "#/components/ui/label";
import { updateSettings } from "#/endpoints/settings";
import { RelationshipLabel } from "#/features/people/relationship-label";
import {
	getUserPrefs,
	type RelationshipDisplay,
	saveUserPrefs,
} from "#/lib/format-datetime";
import type { UserSettings } from "#/schemas/settings";

const RELATIONSHIP_DISPLAY_OPTIONS: {
	value: RelationshipDisplay;
	label: string;
}[] = [
	{ value: "plain", label: "Type, then name (default)" },
	{ value: "sentence", label: "Sentence" },
	{ value: "badge", label: "Name with badge" },
];

// Minh has Lan as em ruột, so Lan's own page lists Minh as anh ruột.
const EXAMPLE_PROFILES = [
	{ self: "Minh", other: "Lan", type: "em ruột", inverse: "anh ruột" },
	{ self: "Lan", other: "Minh", type: "anh ruột", inverse: "em ruột" },
];

export function RelationshipDisplaySettingsCard({
	apiSettings,
	isPlaceholderData,
	buildPayload,
	queryClient,
}: {
	apiSettings: UserSettings | undefined;
	isPlaceholderData: boolean;
	buildPayload: (
		overrides?: Partial<Parameters<typeof updateSettings>[0]>,
	) => Parameters<typeof updateSettings>[0];
	queryClient: ReturnType<typeof useQueryClient>;
}) {
	const [display, setDisplay] = useState<RelationshipDisplay>(
		() => getUserPrefs().relationshipDisplay,
	);

	const [synced, setSynced] = useState(false);
	if (apiSettings && !isPlaceholderData && !synced) {
		setDisplay(apiSettings.relationship_display);
		setSynced(true);
	}

	const mutation = useMutation({
		mutationFn: () =>
			updateSettings(buildPayload({ relationship_display: display })),
		onSuccess: (updated) => {
			saveUserPrefs({
				relationshipDisplay:
					updated.relationship_display as RelationshipDisplay,
			});
			queryClient.setQueryData(["settings"], updated);
		},
	});

	return (
		<Card>
			<CardHeader>
				<CardTitle className="text-[14px] font-medium text-ink">
					Relationships
				</CardTitle>
				<CardDescription className="text-[12px] text-sub">
					How relationship rows read on a person's profile. The type always
					describes the other person: on Minh's profile, "em ruột Lan" means Lan
					is Minh's em ruột.
				</CardDescription>
			</CardHeader>
			<CardContent className="space-y-4">
				<div className="space-y-2">
					<Label className="text-[13px]">Display style</Label>
					<div className="space-y-1.5">
						{RELATIONSHIP_DISPLAY_OPTIONS.map((opt) => (
							<label
								key={opt.value}
								className="flex items-center gap-3 cursor-pointer"
							>
								<input
									type="radio"
									name="relationshipDisplay"
									value={opt.value}
									checked={display === opt.value}
									onChange={() => setDisplay(opt.value)}
									className="accent-accent"
								/>
								<span className="text-[13px] text-ink">{opt.label}</span>
							</label>
						))}
					</div>
				</div>

				<div className="space-y-2">
					<Label className="text-[13px]">Example</Label>
					<div className="grid gap-2 sm:grid-cols-2">
						{EXAMPLE_PROFILES.map((ex) => (
							<div
								key={ex.self}
								className="border-bw border-line rounded-md p-2 space-y-1.5"
							>
								<p className="text-[11px] text-sub">On {ex.self}'s profile</p>
								<div className="flex items-center gap-3 text-sm">
									<RelationshipLabel
										display={display}
										typeName={ex.type}
										inverseTypeName={ex.inverse}
										selfName={ex.self}
										otherName={ex.other}
										other={<span className="font-medium">{ex.other}</span>}
									/>
								</div>
							</div>
						))}
					</div>
				</div>

				<Button
					onClick={() => mutation.mutate()}
					size="sm"
					disabled={isPlaceholderData || mutation.isPending}
				>
					{mutation.isPending
						? "Saving…"
						: mutation.isSuccess
							? "Saved!"
							: "Save display style"}
				</Button>
				{mutation.isError && (
					<p className="text-[12px] text-danger-fg">
						Failed to save. Please try again.
					</p>
				)}
			</CardContent>
		</Card>
	);
}
