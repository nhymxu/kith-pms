<!-- intent-skills:start -->
## Skill Loading

Before substantial work:
- Skill check: run `npx @tanstack/intent@latest list`, or use skills already listed in context.
- Skill guidance: if one local skill clearly matches the task, run `npx @tanstack/intent@latest load <package>#<skill>` and follow the returned `SKILL.md`.
- Monorepos: when working across packages, run the skill check from the workspace root and prefer the local skill for the package being changed.
- Multiple matches: prefer the most specific local skill for the package or concern you are changing; load additional skills only when the task spans multiple packages or concerns.
<!-- intent-skills:end -->

Read before starting:
- `README.md` — features, setup, env vars, `make` targets, project layout.
- `docs/code-standards.md` — Go/React conventions, bun ORM and repo patterns.

Stack versions live in `go.mod` and `web/package.json`. Read them; do not trust a copy.

This file records only what those sources do **not** say: invariants that span files, and edges that have already caused bugs.

---

## Code Style

- **No pointless comments**: comment only non-obvious *why*. Never describe what the code does.

### Commit messages

- Conventional Commits: `<type>(<scope>): <description>`.
- Title ≤50 chars (max 72), imperative, no period. Summarize the objective — no file, variable, or line-count details.
- Body: max 5 bullets, <20 words each, explaining *why*. No prose.

## Auth contract

- Session is the `kith_session` **HttpOnly cookie** set by `POST /v1/auth/login`.
- Every mutating call (`POST/PUT/PATCH/DELETE /v1/*`) must send `X-Requested-With: kith-spa`. This *is* the CSRF defence — a custom header cross-origin attackers cannot set without a preflight the server rejects. Applies to `FormData`/multipart requests too (avatar, gift images), which is the case most often forgotten.
- `GET /v1/auth/me` returns `{user}` or 401 — initialises auth state on load.
- `TOKEN_AUTH` is **server-side only**. Never reference it in frontend code or `web/.env*`.

## Frontend

- **Path alias**: `#/`, not `@/` — mapped in `web/package.json` `imports` (`"#/*": "./src/*"`).
- **Semantic tokens only**: style with the token classes defined in `web/src/styles.css` (`bg-panel`, `text-ink`, `text-sub`, `border-line`, `border-bw`, `bg-chip`, status triads such as `--success-bg/-fg/-line`). Never hardcode Tailwind palette classes (`zinc`, `gray`, `indigo`, …) — they ignore the active theme.
- **Themes are not a dark mode**: 6 themes apply via `html[data-theme]`. `nightdesk` is a dark *theme variant*, so there is no `.dark` class anywhere — do not add one.
- **Form controls**: interactive fields (input, textarea, select trigger, checkbox, radio, switch) use `border-field-bw border-field-line`, never `border-bw border-line` — `--field-line` is tuned to ~3:1 against `--field` so fields stay visible, and `--field-bw` stays 1px in softclay where `--bw` is 0. `border-line` remains correct for cards, dividers, and containers.
- **Dead neobrutalism classes**: the old `--main`, `--main-foreground`, and `--secondary-background` tokens were removed from `styles.css`. Classes like `bg-secondary-background` and `border-main` still linger in a few components (e.g. `web/src/features/people/avatar-uploader.tsx`) and resolve to nothing. Replace them with semantic tokens when touching that code; never add new ones.
- **New UI components**: `web/components.json` is a standard shadcn config (new-york, zinc, `#/` aliases). Prefer the local `shadcn` skill. If a CLI add is blocked in this environment, write the file into `web/src/components/ui/` by hand — then convert palette classes to semantic tokens before committing.
- **No browser Sentry**: `sentry-go` is server-side only. Do not add `@sentry/react` or any browser Sentry SDK to the frontend bundle.

## Known gotchas

- **Settings enums are duplicated across Go and TS.** Adding a theme, nav layout, or search scope means editing every member of its sync set, or zod/Go validation fails silently. The authoritative lists are the comments beside each `ErrInvalid*` sentinel in `internal/settings/service.go` — read them before editing. `theme` has a fourth site the others lack: the FOUC guard in `web/index.html` hardcodes the allow-list alongside `THEMES` in `web/src/lib/theme.ts` and `validThemes` in `internal/settings/service.go`. `number_format` has five: Go `validNumberFormats`, the zod enum, the `NumberFormat` union + `DEFAULTS` in `web/src/lib/format-datetime.ts`, `NUMBER_FORMAT_OPTIONS` in `web/src/routes/_authed/settings/_layout.general.tsx`, and the separator branch in `web/src/lib/format-currency.ts` (`formatNumber`) — that last one is behavioral, not just validation, so a new style can pass validation and still render wrong.
- **SPA is embedded at compile time.** `//go:embed all:public` in `internal/api/spa/spa.go` bakes in `internal/api/spa/public/`. Changes under `web/src/` need `make web` **and** a Go rebuild before they appear in the binary.
- **`internal/api/spa/public/.gitignore` must stay tracked.** `.gitignore` ignores the directory contents but whitelists this one file, so `//go:embed` still finds a non-empty dir on a fresh checkout. Deleting it breaks compilation for everyone.
- **CSP is hardcoded in Go.** `setIndexHeaders` in `internal/api/spa/spa.go` sends `default-src 'self'; script-src 'self'; connect-src 'self'` (plus `img-src` `data:`/`blob:`). Any call to an external host, CDN script, or remote font from the SPA is blocked at runtime with nothing failing at build time. Widening the policy is a deliberate server-side change.
- **Deep-link refresh depends on the catch-all.** `spaFallback` returns `index.html` for every non-API GET; it explicitly excludes `/health`, `/v1/*`, and `/assets/*` so those still 404 naturally. Mount it last, after API and health routes.

## Before completing any task

- [ ] `make lint` passes — runs `lint-go` (golangci-lint), `lint-biome`, and `lint-tsc`.
- [ ] Errors are handled, not swallowed.
- [ ] Tests written and passing: `make tests` (Go, `-race`, integration tag) or `make test-coverage`.
