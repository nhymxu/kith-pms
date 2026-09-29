// Notes endpoints: list-by-person, list-all, get, create, update, delete
import { apiFetch } from "../lib/api-client";
import {
	type Note,
	type NoteAllList,
	type NoteList,
	type NoteRequest,
	noteAllListSchema,
	noteListSchema,
	noteSchema,
} from "../schemas/note";

type Envelope<T> = { data: T };

export interface NoteListParams {
	page?: number;
	page_size?: number;
	from_date?: string;
	to_date?: string;
}

export async function listNotesByPerson(
	personId: number,
	params: NoteListParams = {},
): Promise<NoteList> {
	const qs = new URLSearchParams();
	if (params.page) qs.set("page", String(params.page));
	if (params.page_size) qs.set("page_size", String(params.page_size));
	if (params.from_date) qs.set("from_date", params.from_date);
	if (params.to_date) qs.set("to_date", params.to_date);

	const query = qs.toString();
	const res = await apiFetch<Envelope<unknown>>(
		`/v1/people/${personId}/notes${query ? `?${query}` : ""}`,
	);
	return noteListSchema.parse(res.data);
}

export interface AllNotesParams extends NoteListParams {
	person_ids?: number[];
}

export async function listAllNotes(
	params: AllNotesParams = {},
): Promise<NoteAllList> {
	const qs = new URLSearchParams();
	if (params.page) qs.set("page", String(params.page));
	if (params.page_size) qs.set("page_size", String(params.page_size));

	if (params.person_ids?.length)
		qs.set("person_ids", params.person_ids.join(","));
	if (params.from_date) qs.set("from_date", params.from_date);
	if (params.to_date) qs.set("to_date", params.to_date);

	const query = qs.toString();
	const res = await apiFetch<Envelope<unknown>>(
		`/v1/notes${query ? `?${query}` : ""}`,
	);
	return noteAllListSchema.parse(res.data);
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
