// Notes endpoints: list-by-person, get, create, update, delete
import { apiFetch } from "../lib/api-client";
import {
	type Note,
	type NoteList,
	type NoteRequest,
	noteListSchema,
	noteSchema,
} from "../schemas/note";

type Envelope<T> = { data: T };

export interface NoteListParams {
	page?: number;
	page_size?: number;
}

export async function listNotesByPerson(
	personId: number,
	params: NoteListParams = {},
): Promise<NoteList> {
	const qs = new URLSearchParams();
	if (params.page) qs.set("page", String(params.page));
	if (params.page_size) qs.set("page_size", String(params.page_size));

	const query = qs.toString();
	const res = await apiFetch<Envelope<unknown>>(
		`/v1/people/${personId}/notes${query ? `?${query}` : ""}`,
	);
	return noteListSchema.parse(res.data);
}

export async function createNote(
	personId: number,
	body: NoteRequest,
): Promise<Note> {
	const res = await apiFetch<Envelope<unknown>>(
		`/v1/people/${personId}/notes`,
		{
			method: "POST",
			body: JSON.stringify(body),
		},
	);
	return noteSchema.parse(res.data);
}

export async function updateNote(id: number, body: NoteRequest): Promise<Note> {
	const res = await apiFetch<Envelope<unknown>>(`/v1/notes/${id}`, {
		method: "PUT",
		body: JSON.stringify(body),
	});
	return noteSchema.parse(res.data);
}

export async function deleteNote(id: number): Promise<void> {
	await apiFetch(`/v1/notes/${id}`, { method: "DELETE" });
}
