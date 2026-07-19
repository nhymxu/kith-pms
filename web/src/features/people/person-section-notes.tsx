import { NotesList } from "#/features/notes/notes-list";

function SectionHeading({ children }: { children: React.ReactNode }) {
	return (
		<h2 className="text-[11px] font-semibold uppercase tracking-widest text-sub mb-2">
			{children}
		</h2>
	);
}

interface NotesSectionProps {
	personId: number;
}

export function NotesSection({ personId }: NotesSectionProps) {
	return (
		<div>
			<SectionHeading>Notes</SectionHeading>
			<NotesList personId={personId} />
		</div>
	);
}
