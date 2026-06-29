// Important dates endpoints: list by person, replace for person, upcoming
import { apiFetch } from "../lib/api-client";
import {
	type DatesReplaceRequest,
	type ImportantDateList,
	importantDateListSchema,
	type UpcomingDatesList,
	upcomingDatesListSchema,
} from "../schemas/date";

type Envelope<T> = { data: T };

export async function listDatesByPerson(
	personId: number,
): Promise<ImportantDateList> {
	const res = await apiFetch<Envelope<unknown>>(`/v1/people/${personId}/dates`);
	return importantDateListSchema.parse(res.data);
}

export async function replaceDatesForPerson(
	personId: number,
	body: DatesReplaceRequest,
): Promise<void> {
	await apiFetch(`/v1/people/${personId}/dates`, {
		method: "PUT",
		body: JSON.stringify(body),
	});
}

export async function listUpcomingDates(days = 30): Promise<UpcomingDatesList> {
	const res = await apiFetch<Envelope<unknown>>(
		`/v1/important-dates/upcoming?days=${days}`,
	);
	return upcomingDatesListSchema.parse(res.data);
}
