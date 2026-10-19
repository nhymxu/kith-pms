import { readdirSync, readFileSync } from "node:fs";
import { join } from "node:path";
import { describe, expect, it } from "vitest";

const LOCALES_DIR = join(process.cwd(), "src/locales");
const HANDLER_DIR = join(process.cwd(), "../internal/api/handler");

type Catalog = Record<string, unknown>;

function flatten(obj: Catalog, prefix = ""): Record<string, string> {
	const out: Record<string, string> = {};
	for (const [k, v] of Object.entries(obj)) {
		const key = prefix + k;
		if (v && typeof v === "object")
			Object.assign(out, flatten(v as Catalog, `${key}.`));
		else out[key] = String(v);
	}
	return out;
}

// vi has a single plural form (_other), so _one variants exist only in en.
const baseKey = (k: string) => k.replace(/_(one|other)$/, "");

function vars(s: string): string[] {
	return [...s.matchAll(/\{\{\s*(\w+)\s*\}\}|<(\/?\w+)>/g)]
		.map((m) => m[1] ?? m[2].replace("/", ""))
		.sort();
}

function load(lng: string, ns: string) {
	return flatten(
		JSON.parse(readFileSync(join(LOCALES_DIR, lng, `${ns}.json`), "utf8")),
	);
}

const namespaces = readdirSync(join(LOCALES_DIR, "en"))
	.filter((f) => f.endsWith(".json"))
	.map((f) => f.replace(".json", ""));

describe("catalog parity (en vs vi)", () => {
	it("has the same namespace files", () => {
		const vi = readdirSync(join(LOCALES_DIR, "vi"))
			.filter((f) => f.endsWith(".json"))
			.map((f) => f.replace(".json", ""));
		expect(vi.sort()).toEqual([...namespaces].sort());
	});

	it.each(namespaces)("%s: same keys and placeholders", (ns) => {
		const en = load("en", ns);
		const vi = load("vi", ns);
		const enKeys = [...new Set(Object.keys(en).map(baseKey))].sort();
		const viKeys = [...new Set(Object.keys(vi).map(baseKey))].sort();
		expect(viKeys).toEqual(enKeys);

		for (const [k, v] of Object.entries(en)) {
			const viValue = vi[k] ?? vi[`${baseKey(k)}_other`];
			if (viValue === undefined) continue;
			expect(vars(viValue), `${ns}:${k}`).toEqual(vars(v));
		}
	});
});

describe("server error codes", () => {
	it("every code emitted by Go handlers has an en and vi entry", () => {
		const codes = new Set<string>();
		const files = [
			...readdirSync(HANDLER_DIR)
				.filter((f) => f.endsWith(".go") && !f.endsWith("_test.go"))
				.map((f) => join(HANDLER_DIR, f)),
			join(process.cwd(), "../internal/auth/middleware.go"),
		];
		for (const f of files) {
			const src = readFileSync(f, "utf8");
			for (const m of src.matchAll(
				/(?:apiErr|apiErrParams|jsonErr)\(\s*c,\s*http\.Status\w+,\s*"(\w+)"/g,
			))
				codes.add(m[1]);
			for (const m of src.matchAll(/"code":\s+"(\w+)"/g)) codes.add(m[1]);
			for (const m of src.matchAll(/\{settings\.\w+, "(\w+)"\}/g))
				codes.add(m[1]);
		}

		expect(codes.size).toBeGreaterThan(50);

		for (const lng of ["en", "vi"]) {
			const server = (
				JSON.parse(
					readFileSync(join(LOCALES_DIR, lng, "errors.json"), "utf8"),
				) as { server: Catalog }
			).server;
			const missing = [...codes].filter((c) => !(c in server));
			expect(missing, `${lng} missing`).toEqual([]);
		}
	});
});
