import { render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { RelationshipLabel } from "#/features/people/relationship-label";
import type { Language, RelationshipDisplay } from "#/lib/format-datetime";
import i18n, { setLanguage } from "#/lib/i18n";

const props = {
	typeName: "em ruột",
	inverseTypeName: "anh ruột",
	selfName: "Minh",
	otherName: "Lan",
	other: <span>Lan</span>,
};

function renderLabel(display: RelationshipDisplay) {
	return render(
		<div>
			<RelationshipLabel display={display} {...props} />
		</div>,
	);
}

afterEach(() => setLanguage("en"));

describe("RelationshipLabel", () => {
	it.each<[Language, string]>([
		["en", "is em ruột of Minh"],
		["vi", "là em ruột của Minh"],
	])("sentence style in %s", async (lng, text) => {
		await setLanguage(lng);
		renderLabel("sentence");
		expect(screen.getByText(text)).toBeTruthy();
	});

	it.each<[Language, string]>([
		["en", "Show relationship direction"],
		["vi", "Xem chiều quan hệ"],
	])("badge style in %s", async (lng, label) => {
		await setLanguage(lng);
		renderLabel("badge");
		expect(screen.getByLabelText(label)).toBeTruthy();
		expect(screen.getByText("em ruột")).toBeTruthy();
	});

	it.each<[Language, string]>([
		["en", "Show relationship direction"],
		["vi", "Xem chiều quan hệ"],
	])("plain style in %s", async (lng, label) => {
		await setLanguage(lng);
		renderLabel("plain");
		expect(screen.getByLabelText(label)).toBeTruthy();
		expect(screen.getByText("em ruột")).toBeTruthy();
	});

	it("tooltip names both directions in vi", async () => {
		await setLanguage("vi");
		expect(
			i18n.t("people:relationship.tooltip", {
				other: "Lan",
				type: "em ruột",
				self: "Minh",
				inverse: "anh ruột",
			}),
		).toBe("Lan là em ruột của Minh\nMinh là anh ruột của Lan");
	});
});
