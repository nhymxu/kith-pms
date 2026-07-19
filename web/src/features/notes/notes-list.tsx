import {
	useMutation,
	useQueryClient,
	useSuspenseQuery,
} from "@tanstack/react-query";
import { Check, Pencil, Plus, Trash2, X } from "lucide-react";
import { useState } from "react";
import { QueryBoundary } from "#/components/query-boundary";
import { Button } from "#/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import { Input } from "#/components/ui/input";
import { Textarea } from "#/components/ui/textarea";
import {
	createNote,
	deleteNote,
	listNotesByPerson,
	updateNote,
} from "#/endpoints/notes";
import { keys } from "#/query-keys";
import type { Note, NoteRequest } from "#/schemas/note";

interface NoteFormProps {
	note?: Partial<Note>;
	onSave: (body: NoteRequest) => void;
	onCancel: () => void;
	saving: boolean;
}

function NoteForm({ note, onSave, onCancel, saving }: NoteFormProps) {
	const [title, setTitle] = useState(note?.title ?? "");
	const [content, setContent] = useState(note?.content ?? "");

	return (
		<div className="border-bw border-line rounded-md p-3 bg-muted space-y-2">
			<Input
				className="h-8"
				value={title}
				onChange={(e) => setTitle(e.target.value)}
				placeholder="Title (optional)"
			/>
			<Textarea
				rows={3}
				value={content}
				onChange={(e) => setContent(e.target.value)}
				placeholder="Write a note…"
				autoFocus
			/>
			<div className="flex gap-2 justify-end">
				<Button variant="neutral" size="sm" onClick={onCancel}>
					<X className="size-3" /> Cancel
				</Button>
				<Button
					size="sm"
					disabled={!content.trim() || saving}
					onClick={() => onSave({ title: title.trim(), content })}
				>
					<Check className="size-3" /> {saving ? "Saving…" : "Save"}
				</Button>
			</div>
		</div>
	);
}

interface NotesListInnerProps {
	personId: number;
}

function NotesListInner({ personId }: NotesListInnerProps) {
	const qc = useQueryClient();
	const [creating, setCreating] = useState(false);
	const [editingId, setEditingId] = useState<number | null>(null);
	const [confirmDeleteId, setConfirmDeleteId] = useState<number | null>(null);

	const { data } = useSuspenseQuery({
		queryKey: keys.notes.list({ person_id: personId }),
		queryFn: () => listNotesByPerson(personId, { page_size: 200 }),
	});

	function invalidate() {
		qc.invalidateQueries({
			queryKey: keys.notes.list({ person_id: personId }),
		});
	}

	const createMutation = useMutation({
		mutationFn: (body: NoteRequest) => createNote(personId, body),
		onSuccess: () => {
			invalidate();
			setCreating(false);
		},
	});

	const updateMutation = useMutation({
		mutationFn: ({ id, body }: { id: number; body: NoteRequest }) =>
			updateNote(id, body),
		onSuccess: () => {
			invalidate();
			setEditingId(null);
		},
	});

	const deleteMutation = useMutation({
		mutationFn: (id: number) => deleteNote(id),
		onSuccess: () => {
			invalidate();
			setConfirmDeleteId(null);
		},
	});

	const items = data.items;
	const deleteTarget = items.find((n) => n.id === confirmDeleteId);

	return (
		<div>
			<div className="flex items-center justify-end mb-2">
				<Button variant="neutral" size="sm" onClick={() => setCreating(true)}>
					<Plus className="size-3" /> Add note
				</Button>
			</div>

			<div className="space-y-3">
				{creating && (
					<NoteForm
						onSave={(body) => createMutation.mutate(body)}
						onCancel={() => setCreating(false)}
						saving={createMutation.isPending}
					/>
				)}

				{items.map((n) =>
					editingId === n.id ? (
						<NoteForm
							key={n.id}
							note={n}
							onSave={(body) => updateMutation.mutate({ id: n.id, body })}
							onCancel={() => setEditingId(null)}
							saving={updateMutation.isPending}
						/>
					) : (
						<div
							key={n.id}
							className="text-sm border-bw border-line rounded-md p-3 space-y-1"
						>
							<div className="flex items-start gap-2">
								<div className="flex-1 min-w-0">
									{n.title && <p className="font-medium">{n.title}</p>}
									{n.content && (
										<p className="text-sub whitespace-pre-wrap">{n.content}</p>
									)}
								</div>
								<button
									type="button"
									onClick={() => setEditingId(n.id)}
									className="text-foreground/40 hover:text-main shrink-0"
								>
									<Pencil className="size-3" />
								</button>
								<button
									type="button"
									onClick={() => setConfirmDeleteId(n.id)}
									className="text-foreground/40 hover:text-destructive shrink-0"
								>
									<Trash2 className="size-3" />
								</button>
							</div>
							<p className="font-mono text-[11px] text-sub">
								{n.created_at.slice(0, 10)}
							</p>
						</div>
					),
				)}

				{items.length === 0 && !creating && (
					<p className="text-sm text-sub">No notes yet.</p>
				)}
			</div>

			<Dialog
				open={confirmDeleteId !== null}
				onOpenChange={(v) => !v && setConfirmDeleteId(null)}
			>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Delete note?</DialogTitle>
					</DialogHeader>
					{deleteTarget && (
						<p className="text-[13px] text-sub">
							{deleteTarget.title || deleteTarget.content.slice(0, 60)}
						</p>
					)}
					<DialogFooter>
						<Button variant="neutral" onClick={() => setConfirmDeleteId(null)}>
							Cancel
						</Button>
						<Button
							variant="destructive"
							disabled={deleteMutation.isPending}
							onClick={() => {
								if (confirmDeleteId !== null)
									deleteMutation.mutate(confirmDeleteId);
							}}
						>
							{deleteMutation.isPending ? "Deleting…" : "Delete"}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</div>
	);
}

interface NotesListProps {
	personId: number;
}

// Shared list+editor UI mounted from both the self Notes page and the person
// detail Notes section — same owner-scoped query, different person_id.
export function NotesList({ personId }: NotesListProps) {
	return (
		<QueryBoundary>
			<NotesListInner personId={personId} />
		</QueryBoundary>
	);
}
