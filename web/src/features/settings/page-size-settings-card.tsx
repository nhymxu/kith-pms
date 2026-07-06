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
import type { UserSettings } from "#/schemas/settings";

const PAGE_SIZE_OPTIONS = [10, 25, 50, 100, 200] as const;

export function PageSizeSettingsCard({
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
	const [defaultPageSize, setDefaultPageSize] = useState(25);

	const [synced, setSynced] = useState(false);
	if (apiSettings && !isPlaceholderData && !synced) {
		setDefaultPageSize(apiSettings.default_page_size);
		setSynced(true);
	}

	const pageSizeMutation = useMutation({
		mutationFn: () =>
			updateSettings(buildPayload({ default_page_size: defaultPageSize })),
		onSuccess: (updated) => {
			queryClient.setQueryData(["settings"], updated);
		},
	});

	return (
		<Card>
			<CardHeader>
				<CardTitle className="text-[14px] font-medium text-zinc-900">
					Page size
				</CardTitle>
				<CardDescription className="text-[12px] text-zinc-500">
					Default rows per page for the Journal, People, and Audit lists.
					Applies unless a page size is set explicitly in the URL.
				</CardDescription>
			</CardHeader>
			<CardContent className="space-y-4">
				<div className="space-y-2">
					<Label className="text-[13px]">Rows per page</Label>
					<div className="space-y-1.5">
						{PAGE_SIZE_OPTIONS.map((size) => (
							<label
								key={size}
								className="flex items-center gap-3 cursor-pointer"
							>
								<input
									type="radio"
									name="defaultPageSize"
									value={size}
									checked={defaultPageSize === size}
									onChange={() => setDefaultPageSize(size)}
									className="accent-indigo-600"
								/>
								<span className="text-[13px] text-zinc-700">{size}</span>
							</label>
						))}
					</div>
				</div>

				<Button
					onClick={() => pageSizeMutation.mutate()}
					size="sm"
					disabled={pageSizeMutation.isPending}
				>
					{pageSizeMutation.isPending
						? "Saving…"
						: pageSizeMutation.isSuccess
							? "Saved!"
							: "Save default"}
				</Button>
				{pageSizeMutation.isError && (
					<p className="text-[12px] text-red-500">
						Failed to save. Please try again.
					</p>
				)}
			</CardContent>
		</Card>
	);
}
