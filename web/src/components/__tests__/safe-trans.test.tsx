import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { SafeTrans } from "#/components/safe-trans";

describe("SafeTrans", () => {
	it("renders markup-like user values literally", () => {
		const name = "<b>evil</b> & <0>a</0>";
		const { container } = render(
			<p>
				<SafeTrans
					ns="people"
					i18nKey="sections.labels.removeBody"
					values={{ name }}
					components={{ b: <strong /> }}
				/>
			</p>,
		);
		expect(container.textContent).toContain(name);
		expect(container.querySelectorAll("strong")).toHaveLength(1);
	});
});
