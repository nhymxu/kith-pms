// Gifts endpoints: list, get, create, update, delete
import { apiFetch } from "../lib/api-client"
import {
	giftWithPersonSchema,
	giftListSchema,
	type GiftWithPerson,
	type GiftList,
	type GiftRequest,
} from "../schemas/gift"

type Envelope<T> = { data: T }

export interface GiftListParams {
	person_id?: number
	page?: number
	page_size?: number
}

export async function listGifts(params: GiftListParams = {}): Promise<GiftList> {
	const qs = new URLSearchParams()
	if (params.person_id) qs.set("person_id", String(params.person_id))
	if (params.page) qs.set("page", String(params.page))
	if (params.page_size) qs.set("page_size", String(params.page_size))

	const query = qs.toString()
	const res = await apiFetch<Envelope<unknown>>(`/v1/gifts${query ? `?${query}` : ""}`)
	return giftListSchema.parse(res.data)
}

export async function getGift(id: number): Promise<GiftWithPerson> {
	const res = await apiFetch<Envelope<unknown>>(`/v1/gifts/${id}`)
	return giftWithPersonSchema.parse(res.data)
}

export async function createGift(body: GiftRequest): Promise<number> {
	const res = await apiFetch<Envelope<{ id: number }>>("/v1/gifts", {
		method: "POST",
		body: JSON.stringify(body),
	})
	return res.data.id
}

export async function updateGift(id: number, body: GiftRequest): Promise<void> {
	await apiFetch(`/v1/gifts/${id}`, {
		method: "PUT",
		body: JSON.stringify(body),
	})
}

export async function deleteGift(id: number): Promise<void> {
	await apiFetch(`/v1/gifts/${id}`, { method: "DELETE" })
}

export async function uploadGiftImage(id: number, file: File): Promise<void> {
	const form = new FormData()
	form.append("image", file)
	await apiFetch(`/v1/gifts/${id}/image`, { method: "POST", body: form })
}

export async function deleteGiftImage(id: number): Promise<void> {
	await apiFetch(`/v1/gifts/${id}/image`, { method: "DELETE" })
}
