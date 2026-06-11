import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Plus, Trash2 } from "lucide-react";
import { useState } from "react";
import { Badge } from "#/components/ui/badge";
import { Button } from "#/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import { attachLabel, detachLabel } from "#/endpoints/people";
import { listPeopleLabels } from "#/endpoints/people-labels";
import { keys } from "#/query-keys";
import type { Person } from "#/schemas/person";

function SectionHeading({ children }: { children: React.ReactNode }) {
	return (
		<h2 className="text-[11px] font-semibold uppercase tracking-widest text-zinc-400 mb-2">
			{children}
		</h2>
	);
}

interface LabelsSectionProps {
	person: Person;
}

export function LabelsSection({ person }: LabelsSectionProps) {
	const qc = useQueryClient();
	const [confirmDetachId, setConfirmDetachId] = useState<number | null>(null);
	const { data: allLabels } = useQuery({
		queryKey: keys.peopleLabels.list(),
		queryFn: listPeopleLabels,
	});
	const attached = person.labels ?? [];
	const attachedIds = new Set(attached.map((l) => l.id));

	const attach = useMutation({
		mutationFn: (labelId: number) => attachLabel(person.id, labelId),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.people.detail(person.id) });
			qc.invalidateQueries({ queryKey: keys.peopleLabels.all });
		},
	});
	const detach = useMutation({
		mutationFn: (labelId: number) => detachLabel(person.id, labelId),
		onSuccess: () => {
			qc.invalidateQueries({ queryKey: keys.people.detail(person.id) });
			qc.invalidateQueries({ queryKey: keys.peopleLabels.all });
		},
	});

	const available = allLabels?.filter((l) => !attachedIds.has(l.id)) ?? [];

	return (
		<div>
			<SectionHeading>Labels</SectionHeading>
			<div className="space-y-3">
				<div>
					<p className="text-[11px] text-zinc-400 mb-1">Attached</p>
					<div className="flex flex-wrap gap-2">
						{attached.map((l) => (
							<div key={l.id} className="flex items-center gap-1">
								<Badge style={{ borderColor: l.color }}>{l.name}</Badge>
								<button
									type="button"
									className="text-foreground/40 hover:text-destructive"
									onClick={() => setConfirmDetachId(l.id)}
								>
									<Trash2 className="size-3" />
								</button>
							</div>
						))}
						{attached.length === 0 && (
							<p className="text-sm text-zinc-400">None attached.</p>
						)}
					</div>
				</div>
				{available.length > 0 && (
					<div>
						<p className="text-[11px] text-zinc-400 mb-1">Available</p>
						<div className="flex flex-wrap gap-2">
							{available.map((l) => (
								<button
									key={l.id}
									type="button"
									onClick={() => attach.mutate(l.id)}
									className="flex items-center gap-1 text-xs border border-dashed border-zinc-300 rounded-md px-2 py-1 hover:border-main transition-colors"
								>
									<Plus className="size-3" />
									{l.name}
								</button>
							))}
						</div>
					</div>
				)}
			</div>
			<Dialog
				open={confirmDetachId !== null}
				onOpenChange={(v) => !v && setConfirmDetachId(null)}
			>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Remove label?</DialogTitle>
					</DialogHeader>
					{(() => {
						const l = attached.find((l) => l.id === confirmDetachId);
						return l ? (
							<p className="text-[13px] text-zinc-600">
								Remove the label <span className="font-medium">{l.name}</span>?
							</p>
						) : null;
					})()}
					<DialogFooter>
						<Button variant="neutral" onClick={() => setConfirmDetachId(null)}>
							Cancel
						</Button>
						<Button
							variant="destructive"
							disabled={detach.isPending}
							onClick={() => {
								if (confirmDetachId !== null) {
									detach.mutate(confirmDetachId);
									setConfirmDetachId(null);
								}
							}}
						>
							{detach.isPending ? "Removing…" : "Remove"}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</div>
	);
}
