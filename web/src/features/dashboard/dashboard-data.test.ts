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
		});
		expect(viewModel.pulse).toHaveLength(14);
	});

	it("classifies reminders, gift ideas, activity, and upcoming moments", () => {
		const source: DashboardSource = {
			people: {
				items: [
					{
						id: 1,
						prefix: "",
						name: "Alex Kim",
						nickname: "",
						relationship_type: "friend",
						other_notes: "",
						avatar_path: "",
						avatar_mime_type: "",
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
						people: [{ person_id: 1, name: "Alex Kim" }],
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
						image_mime_type: "",
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
});
