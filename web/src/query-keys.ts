// Centralised TanStack Query key factory.
// Pattern: keys.<domain>.all → keys.<domain>.list(filters) → keys.<domain>.detail(id)

export type PeopleFilters = {
	q?: string;
	page?: number;
	page_size?: number;
	labels?: number[];
	has_journal?: boolean;
	sort?: string;
};

export type JournalFilters = {
	person_id?: number;
	person_ids?: number[];
	page?: number;
	page_size?: number;
	from_date?: string;
	to_date?: string;
};

export type GiftFilters = {
	person_id?: number;
	page?: number;
	page_size?: number;
};

export type ReminderFilters = {
	person_id?: number;
	completed?: boolean;
	status?: "upcoming" | "overdue" | "all";
};

export type AuditFilters = {
	entity_type?: string;
	entity_id?: number;
	page?: number;
	from_date?: string;
	to_date?: string;
};

export const keys = {
	people: {
		all: ["people"] as const,
		list: (filters: PeopleFilters = {}) => ["people", "list", filters] as const,
		detail: (id: number) => ["people", "detail", id] as const,
		avatar: (id: number) => ["people", "avatar", id] as const,
		relationships: (id: number) => ["people", "relationships", id] as const,
		labels: (id: number) => ["people", "labels", id] as const,
		workHistory: (id: number) => ["people", "work-history", id] as const,
	},
	journal: {
		all: ["journal"] as const,
		list: (filters: JournalFilters = {}) =>
			["journal", "list", filters] as const,
		detail: (id: number) => ["journal", "detail", id] as const,
	},
	gifts: {
		all: ["gifts"] as const,
		list: (filters: GiftFilters = {}) => ["gifts", "list", filters] as const,
		detail: (id: number) => ["gifts", "detail", id] as const,
	},
	reminders: {
		all: ["reminders"] as const,
		list: (filters: ReminderFilters = {}) =>
			["reminders", "list", filters] as const,
		detail: (id: number) => ["reminders", "detail", id] as const,
	},
	dates: {
		all: ["dates"] as const,
		list: (personId: number) => ["dates", "list", personId] as const,
		upcoming: () => ["dates", "upcoming"] as const,
	},
	peopleLabels: {
		all: ["people-labels"] as const,
		list: () => ["people-labels", "list"] as const,
		detail: (id: number) => ["people-labels", "detail", id] as const,
	},
	journalLabels: {
		all: ["journal-labels"] as const,
		list: () => ["journal-labels", "list"] as const,
		detail: (id: number) => ["journal-labels", "detail", id] as const,
	},
	relationshipTypes: {
		all: ["relationship-types"] as const,
		list: () => ["relationship-types", "list"] as const,
		detail: (id: number) => ["relationship-types", "detail", id] as const,
	},
	relationships: {
		graph: (personId?: number) =>
			["relationships", "graph", personId ?? "all"] as const,
	},
	audit: {
		all: ["audit"] as const,
		list: (filters: AuditFilters = {}) => ["audit", "list", filters] as const,
	},
	me: {
		all: ["me"] as const,
		profile: () => ["me", "profile"] as const,
		auth: () => ["me", "auth"] as const,
	},
	app: {
		all: ["app"] as const,
		info: () => ["app", "info"] as const,
	},
} as const;
