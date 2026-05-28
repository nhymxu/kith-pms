# Project Overview — kith-pms

## Vision

kith (kith and kin) is a self-hosted Personal Management System for individuals who want to track relationships, record life events, and retain personal memory — without the overhead of CRM tools designed for sales teams or enterprise workflows.

No collaboration features. No sharing. Just one person's data about the people and moments that matter to them.

## Target Users

Single individual user (self-hosted or personal deployment). No multi-tenancy in scope.

## Core Feature Areas (Implemented)

### Contacts & Relationships
- Store people: name, date of birth, relationship type, contact methods, addresses
- Self-profile designation: mark one person as "Me" for personal journal filtering
- Track interaction history via journal entries linked to people
- Tag relationships with color-coded labels
- Upload profile avatars (JPEG, PNG, GIF, WebP)
- Important dates tracking (birthdays, anniversaries, milestones)
- Reminders for follow-up and staying in touch
- Many-to-many person-to-person relationships with customizable, optionally-paired types
- Last contact timestamp tracking (manual & auto-update from journal entries)

### Life Journal / Log
- Date-stamped entries with title and content
- Link journal entries to multiple contacts
- Full-text search via SQLite FTS5
- Filter by date range and people
- Auto-update last contact for participants when self-profile is involved

### Memory & Notes
- Free-form journal entries tied to people
- Important dates with recurring support
- Searchable via FTS5 full-text index

### Timeline & Reminders
- "On this day" widget showing upcoming important dates
- Reminder system with due dates and completion tracking
- Link reminders to people or important dates

### Gifts & Money Tracking
- Track gifts given, received, and planned
- Money tracking per gift
- Gift photos/images with upload support
- Debt type tracking (owed/owe)

### Audit Log
- Automatic change tracking for all entities
- Timestamps and user attribution
- Full historical record of edits and deletions
- Configurable retention policy with manual cleanup endpoint

### Recurring Reminders
- 7 recurrence types: daily, weekly, monthly, yearly, custom interval, day-of-week, relative-to-last-contact
- Auto-spawn next occurrence when reminder marked complete
- Optional end date cutoff to prevent spawning after specified date

### Goreleaser Multi-Platform Releases
- Automated binary builds for Linux, macOS, Windows (amd64, arm64)
- GitHub Actions CI/CD integration for release automation

## Tech Stack (Implemented)

| Layer            | Technology                                  | Rationale                                                                |
|------------------|---------------------------------------------|--------------------------------------------------------------------------|
| Language         | Go 1.26.2, CGO_ENABLED=0                    | Compiled binary, low overhead, easy self-hosting                         |
| HTTP             | Echo v5.1.1                                 | Lightweight HTTP framework, minimal magic                                |
| Database         | SQLite (modernc.org/sqlite v1.50.1)         | Pure Go, no CGO, single-file database, WAL mode                          |
| ORM              | uptrace/bun v1.2.18                         | Lightweight query builder; raw SQL queries retained; no model layer      |
| Frontend         | React 19, TanStack Router v1                | CSR SPA with file-based routing; full client-side interactivity          |
| Data Fetching    | TanStack Query v5                           | Cache-first data fetching, stale-while-revalidate, per-component refresh |
| Forms            | TanStack Form v0                            | Uncontrolled form state with Zod validation                              |
| Tables           | TanStack Table v8                           | Headless table library for data-heavy views                              |
| UI Components    | Local shadcn-style primitives + Base UI     | Accessible local component APIs with Tailwind theming                    |
| Styling          | Tailwind CSS v4                             | Utility-first CSS with Mist/Blue design tokens                           |
| Build            | Vite 8                                      | Fast bundler; code splitting, lazy loading, HMR                          |
| Linter/Formatter | Biome 2.4.5                                 | Rust-based linter + formatter for JS/TS                                  |
| Package Manager  | pnpm 11                                     | Fast, disk-efficient workspaces                                          |
| CLI              | urfave/cli v3                               | Simple CLI scaffolding for subcommands                                   |
| Config           | koanf v2                                    | Layered config: defaults → .env file → env vars                          |
| Logging          | slog                                        | Standard library structured logging (Go 1.21+)                           |
| Error Monitoring | slog-sentry                                 | Fan-out errors to Sentry without replacing slog                          |
| Auth             | bcrypt + HMAC sessions                      | Password hashing + signed HttpOnly cookie sessions                       |
| Search           | SQLite FTS5                                 | Full-text search with auto-update triggers                               |
| Charts           | Recharts v3.8.1                             | Interactive dashboard visualizations                                     |

## Design System

Linear/Stripe minimal aesthetic: indigo-600 (#4f46e5) accent, zinc surfaces, Inter + JetBrains Mono typography, hairline borders, no shadows, responsive horizontal topbar. Built with React 19 CSR SPA and shadcn/ui components, styled via Tailwind CSS v4 design tokens.

## Deployment & Self-Hosting

### Single Static Binary
- Compiled with `CGO_ENABLED=0` — no runtime dependencies, runs on any Linux/macOS/Windows
- Embedded React SPA (Vite build output compiled into binary)
- Embedded SQL migrations
- All assets (CSS, JS, images) bundled; no external file dependencies

### Docker Deployment
- Multi-stage Dockerfile: Node.js (build SPA) → Go (compile binary) → distroless (runtime)
- `docker-compose.yml` for local self-hosting with volume mounts for data persistence
- Automatic migrations on startup (configurable via `DB_AUTO_MIGRATE`)

### Data Storage
- SQLite database: `data/kith.db` (configurable via `DB_PATH`)
- Avatar storage: `data/avatars/` (configurable via `AVATAR_STORAGE_PATH`)
- Gift images: `data/gifts/` (configurable via `GIFT_STORAGE_PATH`)
- All paths support relative or absolute paths

### Backup & Restore
- `kith-pms backup --to <path>` — SQLite VACUUM INTO (safe while server running)
- `kith-pms restore --from <path>` — restore from backup with safety checks
- No external backup services required

## Security Model

### Authentication
- Single-user password-based authentication (bcrypt hashing)
- HMAC-SHA256 session tokens stored server-side
- HttpOnly SameSite=Lax cookies (Secure when behind TLS)
- Session lifetime: 30 days (configurable via `SESSION_LIFETIME`)

### CSRF Protection
- All state-changing requests (POST/PUT/PATCH/DELETE) require `X-Requested-With: kith-spa` header
- Automatic validation in middleware

### Rate Limiting
- Login attempts: 5 per 15 minutes per IP
- Password changes: 5 per 15 minutes per user

### Data Privacy
- No external services required (self-hosted only)
- Optional Sentry integration for error monitoring (configurable via `SENTRY_DSN`)
- All data remains on user's infrastructure

## Monitoring & Observability

### Logging
- Structured logging via Go's `log/slog` (text or JSON format)
- Request ID tracking for correlation
- Optional Sentry integration for error reporting

### Health Checks
- `GET /health` endpoint (no authentication required) — liveness probe
- `GET /ready` endpoint (no authentication required) — readiness probe (DB pingable + migrations applied)
- Suitable for container orchestration and monitoring

### Metrics
- `GET /metrics` endpoint (Prometheus exposition format, no auth)
- HTTP request metrics: count, duration by method/route/status
- App metrics: DB size, active sessions, build info

## Non-Goals

- Multi-user / team features
- Mobile-native apps (web-first)
- Integration with external CRMs or calendars (at this stage)
- AI features (not in scope for initial phases)
