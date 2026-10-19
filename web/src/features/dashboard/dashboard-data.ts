import i18n from "#/lib/i18n";
import type { AuditEntry } from "#/schemas/audit";
import type { UpcomingDateItem } from "#/schemas/date";
import type { GiftWithPerson } from "#/schemas/gift";
import type { JournalActivity } from "#/schemas/journal";
import type { Person, PersonList } from "#/schemas/person";
import type { ReminderWithPerson } from "#/schemas/reminder";

export type DashboardSource = {
	people?: PersonList;
	archivedPeopleCount?: number;
	favoritePeople?: PersonList;
	lastContactedPeople?: PersonList;
	journal?: { items: JournalActivity[]; total: number };
	reminders?: ReminderWithPerson[];
	dates?: UpcomingDateItem[];
	gifts?: { items: GiftWithPerson[]; total: number };
	audit?: AuditEntry[];
	me?: Person;
};

export type DashboardSummaryCard = {
	id: "people" | "followups" | "dates" | "gifts" | "journal";
	value: number;
	detail: string;
};

export type DashboardPulsePoint = {
	date: string;
	entries: number;
	touches: number;
};

export type DashboardAction = {
	id: string;
	type: "overdue" | "today" | "upcoming" | "gift";
	label: string;
	detail: string;
	date: string;
	personName?: string;
};

export type DashboardActivity = {
	id: string;
	title: string;
	detail: string;
	date: string;
	people: string[];
};

export type DashboardMoment = {
	id: string;
	label: string;
	detail: string;
	date: string;
	personName: string;
};

export type DashboardViewModel = {
	lastUpdatedAt: string;
	summaryCards: DashboardSummaryCard[];
	pulse: DashboardPulsePoint[];
	actions: DashboardAction[];
	activities: DashboardActivity[];
	moments: DashboardMoment[];
	favorites: Person[];
	lastContacted: Person[];
	empty: {
		people: boolean;
		activity: boolean;
		actions: boolean;
		moments: boolean;
		favorites: boolean;
		lastContacted: boolean;
	};
};

const DAY_MS = 24 * 60 * 60 * 1000;
const PULSE_DAYS = 14;

export function buildDashboardViewModel(
	source: DashboardSource,
	now = new Date(),
): DashboardViewModel {
	const t = i18n.getFixedT(null, "dashboard");
	const people = source.people?.items ?? [];
	const journalItems = source.journal?.items ?? [];
	const reminders = source.reminders ?? [];
	const dates = source.dates ?? [];
	const gifts = source.gifts?.items ?? [];
	const openReminders = reminders.filter((reminder) => !reminder.completed);
	const plannedGifts = gifts.filter((gift) => gift.direction === "planned");
	const overdueActions = openReminders.filter(
		(reminder) => compareDateOnly(reminder.due_date, now) < 0,
	);
	const todayActions = openReminders.filter(
		(reminder) => compareDateOnly(reminder.due_date, now) === 0,
	);
	// Sourced from a dedicated `favorite_only=true` query so it reflects the
	// whole network's favorites, not just whichever favorites happen to land
	// on the first page of the general people list.
	const favoritePeople = source.favoritePeople?.items ?? [];
	// Sourced from a dedicated `sort=-last_contact` query (backend ORDER BY
	// last_contact_at DESC NULLS LAST) so it reflects the whole network, not
	// just the page of people already loaded for other dashboard widgets.
	const lastContactedPeople = (source.lastContactedPeople?.items ?? []).filter(
		(p) => p.last_contact_at,
	);

	// Archived people are hidden from every other people query, so the
	// dashboard total is the active count plus a dedicated archived_only=true,
	// page_size=1 count rather than a new backend include_archived param.
	const totalPeopleCount =
		(source.people?.total ?? people.length) + (source.archivedPeopleCount ?? 0);

	return {
		lastUpdatedAt: now.toISOString(),
		summaryCards: [
			{
				id: "people",
				value: totalPeopleCount,
				detail: t("cards.peopleDetail", {
					count: peopleWithRecentContact(people, now),
				}),
			},
			{
				id: "followups",
				value: openReminders.length,
				detail: t("cards.followupsDetail", {
					overdue: overdueActions.length,
					today: todayActions.length,
				}),
			},
			{
				id: "dates",
				value: dates.length,
				detail: t("cards.datesDetail"),
			},
			{
				id: "gifts",
				value: plannedGifts.length,
				detail: t("cards.giftsDetail", {
					count: source.gifts?.total ?? gifts.length,
				}),
			},
			{
				id: "journal",
				value: source.journal?.total ?? journalItems.length,
				detail: t("cards.journalDetail", { count: journalItems.length }),
			},
		],
		pulse: buildPulse(journalItems, now),
		actions: buildActions(openReminders, plannedGifts, now),
		activities: journalItems.map((entry) => ({
			id: String(entry.id),
			title: entry.title,
			detail: entry.content || t("activity.journalEntry"),
			date: entry.occurred_at_date,
			people: entry.people.map((person) => person.name),
		})),
		moments: dates.map((date, index) => ({
			id: `${date.person.id}-${date.kind}-${date.next_occurrence}-${index}`,
			label: i18n.t(`people:sections.dates.kind.${date.kind}`, {
				defaultValue: date.kind,
			}),
			detail:
				date.years_since > 0
					? t("moments.years", { count: date.years_since })
					: t("moments.milestone"),
			date: date.next_occurrence,
			personName: date.person.name,
		})),
		favorites: favoritePeople,
		lastContacted: lastContactedPeople,
		empty: {
			people: totalPeopleCount === 0,
			activity: journalItems.length === 0,
			actions: openReminders.length === 0 && plannedGifts.length === 0,
			moments: dates.length === 0,
			favorites: favoritePeople.length === 0,
			lastContacted: lastContactedPeople.length === 0,
		},
	};
}

function buildPulse(
	entries: JournalActivity[],
	now: Date,
): DashboardPulsePoint[] {
	const buckets = new Map<string, DashboardPulsePoint>();

	for (let offset = PULSE_DAYS - 1; offset >= 0; offset -= 1) {
		const date = new Date(now);
		date.setDate(now.getDate() - offset);
		const key = toDateKey(date);
		buckets.set(key, { date: key, entries: 0, touches: 0 });
	}

	for (const entry of entries) {
		const point = buckets.get(entry.occurred_at_date);
		if (!point) continue;
		point.entries += 1;
		point.touches += Math.max(entry.people.length, 1);
	}

	return [...buckets.values()];
}

function buildActions(
	reminders: ReminderWithPerson[],
	gifts: GiftWithPerson[],
	now: Date,
): DashboardAction[] {
	const t = i18n.getFixedT(null, "dashboard");
	const reminderActions = reminders.map((reminder) => ({
		id: `reminder-${reminder.id}`,
		type: classifyReminder(reminder.due_date, now),
		label: reminder.title,
		detail: reminder.notes || t("actions.reminder"),
		date: reminder.due_date,
		personName: reminder.person_name || undefined,
	})) satisfies DashboardAction[];

	const giftActions = gifts
		.filter((gift) => gift.direction === "planned")
		.map((gift) => ({
			id: `gift-${gift.id}`,
			type: "gift" as const,
			label: gift.title,
			detail: t("actions.giftIdea"),
			date: gift.date || gift.created_at.slice(0, 10),
			personName: gift.person_name,
		}));

	return [...reminderActions, ...giftActions].sort(
		(left: DashboardAction, right: DashboardAction) =>
			left.date.localeCompare(right.date),
	);
}

function classifyReminder(date: string, now: Date): DashboardAction["type"] {
	const comparison = compareDateOnly(date, now);
	if (comparison < 0) return "overdue";
	if (comparison === 0) return "today";
	return "upcoming";
}

function peopleWithRecentContact(people: Person[], now: Date): number {
	return people.filter((person) => {
		if (!person.last_contact_at) return false;
		const lastContact = new Date(person.last_contact_at);
		return (
			Number.isFinite(lastContact.getTime()) &&
			now.getTime() - lastContact.getTime() <= 30 * DAY_MS
		);
	}).length;
}

function compareDateOnly(date: string, now: Date): number {
	return date.localeCompare(toDateKey(now));
}

function toDateKey(date: Date): string {
	const year = date.getFullYear();
	const month = String(date.getMonth() + 1).padStart(2, "0");
	const day = String(date.getDate()).padStart(2, "0");
	return `${year}-${month}-${day}`;
}
