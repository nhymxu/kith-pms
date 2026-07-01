import { describe, expect, it } from "vitest";
import type { DashboardSource } from "./dashboard-data";
import { buildDashboardViewModel } from "./dashboard-data";

const now = new Date("2026-05-17T12:00:00.000Z");

describe("buildDashboardViewModel", () => {
	it("builds empty metrics without crashing", () => {
		const viewModel = buildDashboardViewModel({}, now);

		expect(viewModel.summaryCards.map((card) => card.value)).toEqual([
			0, 0, 0, 0, 0,
		]);
		expect(viewModel.empty).toEqual({
			people: true,
			activity: true,
			actions: true,
			moments: true,
			favorites: true,
		});
		expect(viewModel.pulse).toHaveLength(14);
	});

	it("classifies reminders, gift ideas, activity, and upcoming moments", () => {
		const source: DashboardSource = {
			people: {
				items: [
					{
						id: 1,
						is_self: false,
						is_favorite: false,
						has_birthday_reminder: false,
						prefix: "",
						name: "Alex Kim",
						nickname: "",
						gender: "",
						other_notes: "",
						avatar_path: "",
						avatar_size: 0,
						created_at: "2026-01-01T00:00:00Z",
						updated_at: "2026-05-01T00:00:00Z",
						last_contact_at: "2026-05-10T00:00:00Z",
						contacts: [],
						locations: [],
						labels: [],
					},
				],
				total: 1,
				page: 1,
				page_size: 10,
			},
			journal: {
				items: [
					{
						id: 9,
						title: "Coffee catch-up",
						content: "Discussed trip plans",
						occurred_at_date: "2026-05-17",
						occurred_at_time: "09:00",
						created_at: "2026-05-17T09:00:00Z",
						updated_at: "2026-05-17T09:00:00Z",
						people: [
							{ person_id: 1, name: "Alex Kim", nickname: "", avatar_path: "" },
						],
						labels: [],
					},
				],
				total: 1,
			},
			reminders: [
				{
					id: 2,
					title: "Send follow-up",
					notes: "Share photos",
					due_date: "2026-05-16",
					completed: false,
					created_at: "2026-05-01T00:00:00Z",
					updated_at: "2026-05-01T00:00:00Z",
					person_name: "Alex Kim",
				},
				{
					id: 3,
					title: "Call Sam",
					notes: "",
					due_date: "2026-05-17",
					completed: false,
					created_at: "2026-05-01T00:00:00Z",
					updated_at: "2026-05-01T00:00:00Z",
					person_name: "Sam Lee",
				},
			],
			dates: [
				{
					person: { id: 1, name: "Alex Kim" },
					kind: "birthday",
					date_value: "1990-05-20",
					years_since: 36,
					next_occurrence: "2026-05-20",
				},
			],
			gifts: {
				items: [
					{
						id: 4,
						person_id: 1,
						person_name: "Alex Kim",
						title: "Book",
						direction: "planned",
						date: "2026-05-19",
						notes: "",
						currency: "USD",
						debt_type: "",
						image_path: "",
						created_at: "2026-05-01T00:00:00Z",
						updated_at: "2026-05-01T00:00:00Z",
					},
				],
				total: 1,
			},
		};

		const viewModel = buildDashboardViewModel(source, now);

		expect(viewModel.summaryCards.map((card) => card.value)).toEqual([
			1, 2, 1, 1, 1,
		]);
		expect(viewModel.actions.map((action) => action.type)).toEqual([
			"overdue",
			"today",
			"gift",
		]);
		expect(viewModel.activities[0]).toMatchObject({
			title: "Coffee catch-up",
			people: ["Alex Kim"],
		});
		expect(viewModel.moments[0]).toMatchObject({
			personName: "Alex Kim",
			label: "birthday",
		});
		expect(viewModel.pulse.at(-1)).toMatchObject({
			date: "2026-05-17",
			entries: 1,
			touches: 1,
		});
	});

	it("filters favorites from the people list, capped at 5", () => {
		const person = (id: number, name: string, isFavorite: boolean) => ({
			id,
			is_self: false,
			is_favorite: isFavorite,
			has_birthday_reminder: false,
			prefix: "",
			name,
			nickname: "",
			gender: "",
			other_notes: "",
			avatar_path: "",
			avatar_size: 0,
			created_at: "2026-01-01T00:00:00Z",
			updated_at: "2026-05-01T00:00:00Z",
			contacts: [],
			locations: [],
			labels: [],
		});

		const source: DashboardSource = {
			people: {
				items: [
					person(1, "Alex Kim", true),
					person(2, "Bob Lee", false),
					person(3, "Carol Diaz", true),
				],
				total: 3,
				page: 1,
				page_size: 25,
			},
		};

		const viewModel = buildDashboardViewModel(source, now);

		expect(viewModel.favorites.map((p) => p.name)).toEqual([
			"Alex Kim",
			"Carol Diaz",
		]);
		expect(viewModel.empty.favorites).toBe(false);
	});

	it("marks favorites empty when no one is favorited", () => {
		const source: DashboardSource = {
			people: { items: [], total: 0, page: 1, page_size: 25 },
		};

		const viewModel = buildDashboardViewModel(source, now);

		expect(viewModel.favorites).toEqual([]);
		expect(viewModel.empty.favorites).toBe(true);
	});
});
