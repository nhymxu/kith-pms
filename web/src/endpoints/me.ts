// Me endpoints: self-person profile and setup
import { apiFetch } from "../lib/api-client";
import { type Person, personSchema } from "../schemas/person";

type Envelope<T> = { data: T };

export async function getMe(): Promise<Person> {
	const res = await apiFetch<Envelope<unknown>>("/v1/me");
	return personSchema.parse(res.data);
}

export async function setupMe(personId: number): Promise<number> {
	const res = await apiFetch<Envelope<{ person_id: number }>>("/v1/me/setup", {
		method: "POST",
		body: JSON.stringify({ person_id: personId }),
	});
	return res.data.person_id;
}
