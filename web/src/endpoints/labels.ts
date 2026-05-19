// Labels endpoints: list, get, create, update, delete
import { apiFetch } from "../lib/api-client";
import {
	type Label,
	type LabelList,
	type LabelRequest,
	labelListSchema,
	labelSchema,
} from "../schemas/label";

type Envelope<T> = { data: T };

export async function listLabels(): Promise<LabelList> {
	const res = await apiFetch<Envelope<unknown>>("/v1/labels");
	return labelListSchema.parse(res.data);
}

export async function getLabel(id: number): Promise<Label> {
	const res = await apiFetch<Envelope<unknown>>(`/v1/labels/${id}`);
	return labelSchema.parse(res.data);
}

export async function createLabel(body: LabelRequest): Promise<number> {
	const res = await apiFetch<Envelope<{ id: number }>>("/v1/labels", {
		method: "POST",
		body: JSON.stringify(body),
	});
	return res.data.id;
}

export async function updateLabel(
	id: number,
	body: LabelRequest,
): Promise<void> {
	await apiFetch(`/v1/labels/${id}`, {
		method: "PUT",
		body: JSON.stringify(body),
	});
}

export async function deleteLabel(id: number): Promise<void> {
	await apiFetch(`/v1/labels/${id}`, { method: "DELETE" });
}
