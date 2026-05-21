<!-- intent-skills:start -->
## Skill Loading

Before substantial work:
- Skill check: run `npx @tanstack/intent@latest list`, or use skills already listed in context.
- Skill guidance: if one local skill clearly matches the task, run `npx @tanstack/intent@latest load <package>#<skill>` and follow the returned `SKILL.md`.
- Monorepos: when working across packages, run the skill check from the workspace root and prefer the local skill for the package being changed.
- Multiple matches: prefer the most specific local skill for the package or concern you are changing; load additional skills only when the task spans multiple packages or concerns.
<!-- intent-skills:end -->

This file provides guidance to AI when working with code in this repository.

See `README.md` for full stack details, auth contract, frontend conventions, known gotchas, and development workflow.

---

## Code Style

- **No pointless comments**: Only add comments when explaining non-obvious "why" decisions — never describe what the code does.

### Auth contract

- Authentication uses the `kith_session` **HttpOnly session cookie** set by `POST /v1/auth/login`.
- All mutating API calls (`POST/PUT/PATCH/DELETE /v1/*`) must include the header `X-Requested-With: kith-spa` — this is the CSRF protection mechanism (a custom header cross-origin attackers cannot set without a preflight the server rejects).
- `TOKEN_AUTH` is **server-side only** — never expose it in any frontend bundle or `.env` file.
- `GET /v1/auth/me` returns `{user}` or 401 — used to initialise auth state on load.

### Frontend notes

- **Path alias**: use `#/` (not `@/`) — mapped in `package.json` `imports` as `"#/*": "./src/*"`.
- **Dark mode**: dropped for v1 — do not add `.dark` CSS blocks.
- **UI components**: `pnpm dlx shadcn add` is blocked by the Biome hook in this environment. Fetch registry JSON from `https://neobrutalism.dev/r/<name>.json` and write the component file to `src/components/ui/` manually.
- **Neobrutalism tokens**: `--main`, `--main-foreground`, `--secondary-background`, `--border`, `--shadow` (`= 4px 4px 0 0 var(--border)`).
- **Sentry**: server-side only — no `@sentry/react` or browser Sentry SDK in the frontend bundle.

### Known gotchas

- **Deep-link refresh**: the Go catch-all in `spa.go` returns `index.html` for all non-API GET paths — required for CSR routing.
- **SPA embed**: `internal/web/spa/public/` is baked into the binary at compile time; changes to `web/src/` require `make web` + recompile.
- **`placeholder.txt`**: must stay in `internal/web/spa/public/` so `//go:embed all:public` compiles on a fresh checkout.
- **FormData CSRF**: multipart form endpoints (avatar upload) use `FormData`; the `X-Requested-With: kith-spa` header must still be included.

## Stack

| Layer    | Technology                                             |
|----------|--------------------------------------------------------|
| Language | Go 1.26, `CGO_ENABLED=0`                               |
| HTTP     | Echo v5                                                |
| Database | SQLite (modernc.org/sqlite — pure Go)                  |
| Frontend | React 19, TanStack Router/Query/Table/Form             |
| UI       | shadcn/ui registry + Tailwind v4                       |
| Build    | Vite 6, pnpm, Biome (lint/format)                      |
| Auth     | Argon2id password hash, signed HttpOnly cookies + CSRF |
| Search   | SQLite FTS5                                            |
