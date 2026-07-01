import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { Button } from "#/components/ui/button";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "#/components/ui/select";
import { bulkAssignLabel, listPeopleLabels } from "#/endpoints/people-labels";
import { keys } from "#/query-keys";
import { ConnectToPersonDialog } from "./connect-to-person-dialog";

interface BulkActionBarProps {
	selectedCount: number;
	personIds: number[];
	onClear: () => void;
}

export function BulkActionBar({
	selectedCount,
	personIds,
	onClear,
}: BulkActionBarProps) {
	const [labelId, setLabelId] = useState<string>("");
	const [lastResult, setLastResult] = useState<string | null>(null);
	const [connectOpen, setConnectOpen] = useState(false);
	const qc = useQueryClient();

	const { data: labels } = useQuery({
		queryKey: keys.peopleLabels.list(),
		queryFn: listPeopleLabels,
	});

	const mutation = useMutation({
		mutationFn: () => bulkAssignLabel(Number(labelId), personIds),
		onSuccess: (data) => {
			setLastResult(`Assigned to ${data.attached} people`);
			qc.invalidateQueries({ queryKey: keys.people.all });
			qc.invalidateQueries({ queryKey: keys.peopleLabels.all });
			setLabelId("");
			onClear();
		},
		onError: () => setLastResult("Failed to assign label"),
	});

	return (
		<div className="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 flex items-center gap-3 px-4 py-3 bg-white border-2 border-[var(--border)] shadow-[4px_4px_0_0_var(--border)] rounded-md">
			<span className="text-sm font-medium text-zinc-700">
				{selectedCount} selected
			</span>
			{lastResult && (
				<span className="text-xs text-zinc-500">{lastResult}</span>
			)}
			<Select
				value={labelId}
				onValueChange={(v) => {
					setLabelId(v as string);
					setLastResult(null);
				}}
			>
				<SelectTrigger className="h-8 w-40">
					<SelectValue placeholder="Assign label…" />
				</SelectTrigger>
				<SelectContent>
					{labels?.map((l) => (
						<SelectItem key={l.id} value={String(l.id)}>
							{l.name}
						</SelectItem>
					))}
				</SelectContent>
			</Select>
			<Button
				size="sm"
				disabled={!labelId || mutation.isPending}
				onClick={() => mutation.mutate()}
			>
				{mutation.isPending ? "Assigning…" : "Assign"}
			</Button>
			<Button size="sm" variant="neutral" onClick={() => setConnectOpen(true)}>
				Connect to person…
			</Button>
			<Button size="sm" variant="neutral" onClick={onClear}>
				Clear
			</Button>
			<ConnectToPersonDialog
				selectedIds={personIds}
				open={connectOpen}
				onClose={() => setConnectOpen(false)}
				onSuccess={() => {
					setConnectOpen(false);
					onClear();
				}}
			/>
		</div>
	);
}
