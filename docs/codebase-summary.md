# Codebase Summary

## Directory Structure

```
kith-pms/
├── cmd/                          # CLI entrypoints (compiled to single binary)
│   ├── main.go                   # Binary entry, dependency init, subcommand dispatch
│   ├── web_server.go             # `serve` subcommand — starts Echo HTTP server
│   ├── migrate.go                # `migrate` subcommand — runs schema migrations
│   ├── set_password.go           # `set-password` subcommand — interactive password setter
│   ├── backup.go                 # `backup` subcommand — database backup CLI
│   ├── restore.go                # `restore` subcommand — database restore CLI
│   └── monica_import.go          # `monica-import` subcommand — imports Monica PRM JSON exports
├── internal/                     # Private application code
│   ├── db/                       # Database layer
│   │   ├── sqlite.go             # SQLite connection with WAL + ForeignKeys PRAGMAs
│   │   ├── migrations.go         # SQL migration loader and executor
│   │   └── migrations/           # SQL schema files (0001-0015)
│   ├── auth/                     # Authentication & session management
│   │   ├── domain.go             # User, Session, CSRF token data structures
│   │   ├── password.go           # bcrypt hashing & verification
│   │   ├── sessions.go           # HMAC token generation & validation
│   │   ├── csrf.go               # CSRF token generation & validation
│   │   ├── middleware.go         # Echo middleware for auth/session checks
│   │   ├── service.go            # Auth business logic (login, logout, token refresh)
│   │   ├── repo.go               # User & session queries
│   │   ├── service_test.go       # Service unit tests
│   │   └── password_test.go      # Password hashing tests
│   ├── audit/                    # Audit logging & change tracking
│   │   ├── domain.go             # Entry, EntityType, Action structures
│   │   ├── context.go            # Actor context helpers (WithActor, ActorFromCtx)
│   │   ├── service.go            # Audit logging service (best-effort Log)
│   │   ├── repo.go               # Audit log database queries
│   │   └── service_test.go       # Service unit tests
│   ├── people/                   # Contacts & relationships
│   │   ├── domain.go             # Person, Contact, Location structures
│   │   ├── service.go            # CRUD and query business logic; self-profile management
│   │   ├── repo.go               # Database queries; GetSelfPerson, SetSelfPerson
│   │   └── service_test.go       # Service unit tests
│   ├── labels/                   # Tags & categorization
│   │   ├── domain.go             # Label structure
│   │   ├── service.go            # CRUD and association logic
│   │   ├── repo.go               # Database queries
│   │   └── service_test.go       # Service unit tests
│   ├── journal/                  # Activity log & FTS5
│   │   ├── domain.go             # Entry, EntryLink structures
│   │   ├── service.go            # CRUD and search business logic
│   │   ├── repo.go               # Database queries + FTS5 search
│   │   └── service_test.go       # Service unit tests
│   ├── dates/                    # Important dates & milestones
│   │   ├── domain.go             # ImportantDate, OnThisDayItem structures
│   │   ├── service.go            # CRUD and date query business logic
│   │   ├── repo.go               # Database queries for dates
│   │   └── service_test.go       # Service unit tests
│   ├── reminders/                # Reminders & notifications
│   │   ├── domain.go             # Reminder, ReminderWithPerson structures
│   │   ├── service.go            # CRUD and completion tracking business logic
│   │   ├── repo.go               # Database queries for reminders
│   │   └── service_test.go       # Service unit tests
│   ├── gifts/                    # Gift management & debt tracking
│   │   ├── domain.go             # Gift, GiftWithPerson structures; Direction, DebtType constants
│   │   ├── service.go            # CRUD, image upload/delete business logic
│   │   ├── repo.go               # Database queries for gifts
│   │   └── service_test.go       # Service integration tests
│   ├── relationships/            # Person-to-person relationship junctions
│   │   ├── domain.go             # RelationshipType, PersonRelationship, RelationshipView structures
│   │   ├── service.go            # CreateType/UpdateType/DeleteType; AttachRelationship/DetachRelationship/ListByPerson
│   │   ├── repo.go               # sqlRelationshipTypeRepo + sqlPersonRelationshipRepo; paired tx writes
│   │   └── service_test.go       # 10 integration tests (paired rows, self-loop guard, FK restrict)
│   ├── settings/                 # User settings & preferences
│   │   ├── domain.go             # UserSettings struct (date_format, time_format, timezone) + Defaults
│   │   ├── service.go            # Get/Update business logic with validation
│   │   └── repo.go               # Key/value store queries (GetAll, Set)
│   ├── files/                    # File storage service
│   │   ├── service.go            # LocalFileService for avatar uploads
│   │   └── service_test.go       # File service unit tests
│   ├── metrics/                  # Prometheus metrics & observability
│   │   ├── metrics.go            # Registry, HTTP middleware, collectors (DB size, sessions, build info)
│   │   └── metrics_test.go       # Metrics unit tests
│   ├── monica/                   # Monica PRM import package
│   │   ├── parser.go             # Monica JSON export format unmarshaling
│   │   ├── mapper.go             # Field mapping from Monica to kith-pms domain
│   │   └── mapper_test.go        # Unit tests for Monica-to-domain mapping
│   ├── api/                      # HTTP API handlers
│   │   ├── handlers_auth.go      # Authentication endpoints
│   │   ├── handlers_people.go    # People CRUD endpoints
│   │   ├── handlers_journal.go   # Journal endpoints
│   │   ├── handlers_gifts.go     # Gift endpoints with image upload
│   │   ├── handlers_reminders.go # Reminder endpoints
│   │   ├── handlers_dates.go     # Important dates endpoints
│   │   ├── handlers_labels.go    # Label endpoints
│   │   ├── handlers_relationships.go # Relationship type endpoints
│   │   ├── handlers_audit.go     # Audit log endpoints
│   │   └── handlers_me.go        # User profile endpoints
│   └── web/                      # HTTP server & SPA embedding
│       ├── server.go             # Echo setup (global middleware)
│       ├── route.go              # Route mounting: /health, /v1/*, spa.Handler()
│       └── spa/                  # Embedded React SPA
│           ├── spa.go            # //go:embed all:public; Handler() with catch-all fallback
│           └── public/           # Populated by `make web` (gitignored except placeholder.txt)
├── pkg/                          # Shared packages
│   ├── config/                   # Configuration management
│   │   ├── env.go                # LoadConfig(), EnvConfigMap, environment parsing
│   │   ├── const.go              # Application constants
│   │   └── default.go            # Default config values
│   └── errors/                   # Custom error types
├── internal/db/migrations/       # SQL schema files
│   ├── 0001_init.sql             # Initial schema (users, people, labels)
│   ├── 0002_user_session.sql     # Session tokens table
│   ├── 0003_person.sql           # Person table refine (contacts, locations)
│   ├── 0004_label.sql            # Label-person association table
│   ├── 0005_activity.sql         # Journal entries & person linking
│   ├── 0006_activity_fts.sql     # FTS5 virtual table + triggers
│   ├── 0007_important_date.sql   # Important dates table with virtual month_day column
│   ├── 0008_reminder.sql         # Reminders table with person/date associations
│   ├── 0009_person_avatar.sql    # Avatar metadata columns on person table
│   ├── 0010_work_history.sql     # Work history table
│   ├── 0011_audit_log.sql        # Audit log table for entity change tracking
│   ├── 0012_gift.sql             # Gift table with direction, debt type, and image columns
│   ├── 0013_person_self.sql      # is_self column with unique index for self-profile
│   ├── 0014_person_last_contact.sql  # last_contact_at column on person
│   └── 0015_relationship_type.sql   # relationship_type + person_relationship tables
├── scripts/
│   ├── lint.sh                   # Runs golangci-lint
│   ├── dependency-graph.sh       # Generates module dependency graph
│   └── find-cgo-pkg.sh           # Identifies CGO dependencies
├── web/                          # React SPA source (pnpm workspace)
│   ├── src/
│   │   ├── components/           # UI components (app-shell, data-table, form, ui)
│   │   ├── endpoints/            # API client functions (8 domains)
│   │   ├── features/             # Feature components (9 areas)
│   │   ├── lib/                  # api-client.ts, auth-context.tsx, query-client.ts, utils.ts
│   │   ├── routes/               # File-based routing (TanStack Router)
│   │   ├── schemas/              # Zod validation schemas (10 domains)
│   │   ├── query-keys.ts         # Centralized query key factory
│   │   └── styles.css            # Tailwind + design tokens
│   ├── public/                   # Static assets
│   ├── package.json              # pnpm workspace config
│   └── vite.config.ts            # Vite build configuration
├── data/                         # Runtime data (gitignored)
│   ├── kith.db                   # SQLite database
│   ├── avatars/                  # Avatar files
│   └── gifts/                    # Gift images
├── docs/                         # Project documentation
├── Dockerfile                    # Multi-stage container build
├── docker-compose.yml            # Local development setup
├── Makefile                      # Dev workflow targets
├── go.mod / go.sum               # Module definition and lockfile
└── .env.example                  # Example environment variables
```

## Package Responsibilities

### `cmd` (package `main`)
- **main.go**: Binary entry point; initializes CLI app and dispatches to subcommands
- **web_server.go**: `serve` subcommand — starts Echo server after dependency init
- **migrate.go**: `migrate` subcommand — applies pending SQL migrations
- **set_password.go**: `set-password` subcommand — interactive password change
- **backup.go** & **restore.go**: Database backup/restore CLI with safety checks
- **monica_import.go**: `monica-import` subcommand — imports contacts, labels, activities, reminders, and dates from a Monica PRM JSON export

### `internal/db` — Database layer
- **sqlite.go**: Opens SQLite with modernc.org/sqlite (no CGO); applies PRAGMAs for WAL, foreign keys, safe sync
- **migrations.go**: Loads SQL files from `internal/db/migrations/`, executes in order, tracks applied versions

### `internal/auth` — Single-user authentication
- **domain.go**: User, Session, CSRFToken, PasswordReset data structures
- **password.go**: bcrypt hashing (via golang.org/x/crypto) with verification
- **sessions.go**: HMAC-SHA256 token generation (server-signed); validates session ID from DB against signed token
- **csrf.go**: Per-request CSRF token generation & validation
- **middleware.go**: Echo middleware for session validation and CSRF checks
- **service.go**: Login/logout business logic; token refresh; password changes
- **repo.go**: SQL queries for user lookups, session CRUD, token management

### `internal/audit` — Audit logging & change tracking
- **domain.go**: Entry (id, entity_type, entity_id, entity_name, action, actor_id, created_at), EntityType enum, Action enum
- **context.go**: Helper functions for actor context — `WithActor(ctx, userID)` and `ActorFromCtx(ctx)`
- **service.go**: `Log(ctx, entityType, entityID, entityName, action)` — best-effort logging (never blocks, errors logged as warnings)
- **repo.go**: Database queries for audit log insertion and list retrieval with filtering
- **service_test.go**: Tests for logging behavior and list queries

### `internal/people` — Contacts management
- **domain.go**: Person (name, DOB, type, is_self, last_contact_at), Contact (email, phone), Location (street, city, country)
- **service.go**: CRUD (CreatePerson, GetPerson, UpdatePerson, DeletePerson); query by label, search; self-profile management (GetSelfPerson, SetSelfPerson); UpdateLastContact(personID, contactTime) for manual updates
- **repo.go**: Raw database/sql queries; JOIN queries for contacts & locations; self-profile queries; UpdateLastContact for timestamp persistence
- **service_test.go**: Integration tests for CRUD, complex queries, and self-profile operations

### `internal/labels` — Tag system
- **domain.go**: Label (name, color hex)
- **service.go**: CRUD; many-to-many label-person associations
- **repo.go**: Queries for label lists, person-label associations
- **service_test.go**: Integration tests for many-to-many logic

### `internal/journal` — Activity log with full-text search
- **domain.go**: Entry (title, content, date, author), EntryLink (link to People via FK)
- **service.go**: CRUD; full-text search via FTS5; link entries to multiple people; auto-update last_contact_at for participants when self is included and activity date is newer
- **repo.go**: Queries including FTS5 search; maintains FTS5 trigger-based index
- **service_test.go**: Integration tests for FTS5 search

### `internal/dates` — Important dates & milestones
- **domain.go**: ImportantDate (kind, label, date_value, recurring), OnThisDayItem (date + person info)
- **service.go**: CRUD for dates; OnThisDay queries; Upcoming dates calculation
- **repo.go**: Queries for dates by person; OnThisDay matches; virtual month_day column queries
- **service_test.go**: Integration tests for date parsing, recurring logic, and queries

### `internal/reminders` — Reminders & notifications
- **domain.go**: Reminder (title, notes, due_date, person_id, important_date_id, completed, recurrence_rule, recurrence_end_date), ReminderWithPerson; RecurrenceType enum (7 types), RecurrenceRule struct
- **recurrence.go**: Pure `computeNextDue` function computing next due date for all 7 recurrence types (daily, weekly, monthly, yearly, custom, day_of_week, relative_contact)
- **service.go**: CRUD for reminders; completion tracking; auto-spawn next occurrence when RecurrenceRule != nil; JournalLastContacter interface for relative-contact lookups
- **repo.go**: Queries for reminders with person joins; status filtering; read/write recurrence columns (JSON marshal/unmarshal)
- **service_test.go**: Integration tests for reminder CRUD, completion, and auto-spawn with recurrence

### `internal/gifts` — Gift management & debt tracking
- **domain.go**: Gift (title, description, direction, debt_type, person_id, image_path), GiftWithPerson; Direction and DebtType enums
- **service.go**: CRUD for gifts; UploadImage/DeleteImage for gift photos; persists metadata (path, MIME type, size, upload timestamp); GetByIDWithPerson returns gift with person_name
- **repo.go**: Queries for gifts with person joins; UpdateImage metadata updates; GetByIDWithPerson for detail view
- **service_test.go**: Integration tests for gift CRUD and image operations

### `internal/relationships` — Person-to-person relationships
- **domain.go**: RelationshipType (name, reverse_name, optional inverse_type_id), PersonRelationship (from/to person IDs, type, notes), RelationshipView (rendered relationship with resolved type names)
- **service.go**: CreateType/UpdateType/DeleteType for relationship types; AttachRelationship/DetachRelationship for person junctions; handles symmetric and asymmetric paired types with bidirectional row creation
- **repo.go**: sqlRelationshipTypeRepo and sqlPersonRelationshipRepo; paired transaction writes for bidirectional relationships; FindPair for locating inverse rows
- **service_test.go**: 10 integration tests covering paired rows, self-loop guards, FK constraints, symmetric type bidirectionality

### `internal/settings` — User settings & preferences
- **domain.go**: UserSettings (date_format, time_format, timezone) with Defaults constant
- **service.go**: Get/Update business logic with validation for format/timezone values
- **repo.go**: Key/value store queries (GetAll, Set) for user_setting table

### `internal/files` — File storage service
- **service.go**: LocalFileService for avatar uploads with MIME validation, size limits, path traversal prevention
- **service_test.go**: File service unit tests

### `internal/metrics` — Prometheus metrics & observability
- **metrics.go**: Custom Registry with Go runtime + process collectors; HTTP middleware (request count + latency by method/route/status); GaugeFunc collectors for DB size (PRAGMA page_count) and active sessions; build info gauge; promhttp handler for `/metrics` endpoint
- **metrics_test.go**: Unit tests for route-template label cardinality, unknown route handling, scrape format validation

### `internal/monica` — Monica PRM data import
- **parser.go**: Unmarshals Monica JSON export format (contacts, activities, reminders, tags, etc.) into typed structs
- **mapper.go**: Pure-function mapping from Monica domain types to kith-pms domain types (Person, ContactInfo, Location, Activity, Reminder, ImportantDate)
- **mapper_test.go**: Unit tests for edge cases (birthdate year handling, contact type classification, name assembly, tag deduplication)

### `internal/api` — HTTP API handlers
- **handlers_auth.go**: POST /login, /logout, /logout-all, GET /me, POST /password
- **handlers_people.go**: CRUD + avatar upload/delete/get + relationships + labels + dates + work-history + quick journal/gift + last-contact
- **handlers_journal.go**: CRUD with multi-person tagging, search, date range filter
- **handlers_gifts.go**: CRUD + image upload/delete/get, filterable by person/direction/debt_type
- **handlers_reminders.go**: CRUD + PATCH complete, filterable upcoming/overdue/all
- **handlers_dates.go**: GET /upcoming (30-day window)
- **handlers_labels.go**: CRUD with usage counts
- **handlers_relationships.go**: CRUD with usage counts
- **handlers_settings.go**: GET /settings, PUT /settings for user preferences (date format, time format, timezone)
- **handlers_audit.go**: GET with entity_type/entity_id filter, paginated
- **handlers_me.go**: GET profile, POST setup

### `internal/web` — HTTP layer
- **server.go**: Creates Echo instance with global middleware (recover, request ID, gzip, logger, sentry)
- **route.go**: Mounts `/health`, API (`/v1/*` via `api.Mount`), then `spa.Handler()` as catch-all; injects session loader + audit actor middleware
- **spa/spa.go**: Embeds `public/` via `//go:embed all:public`; serves `/assets/*` with 1-year immutable cache; returns `index.html` (no-cache, CSP headers) for all non-API GET paths; real 404 for unknown `/assets/*` paths
- **spa/public/**: Populated at build time by `make web` (copies `web/dist/` here); gitignored except `placeholder.txt` sentinel

## React SPA Frontend (`web/` directory)

**Stack**: React 19 CSR SPA with TanStack Router v1 (file-based routing, not React Router v6), TanStack Query v5, TanStack Table v8, TanStack Form v0, shadcn/ui (Linear/Stripe minimal design), Tailwind CSS v4, Biome 2.4.5, Vite 8, pnpm 11.

**Path Alias**: `#/` (not `@/`) — mapped in `web/package.json` `imports`.

**Build Pipeline**: `pnpm build` → `web/dist/` → copied to `internal/web/spa/public/` via `make web` → embedded in binary via `//go:embed all:public`.

### Directory Structure (`web/src/`)
```
components/ui/           # shadcn components (button, card, input, select, dialog, sheet, etc.)
components/app-shell/    # Layout (topbar.tsx with responsive nav, app-layout.tsx)
endpoints/               # API call functions (per-resource: people.ts, journal.ts, etc.)
features/
  dashboard/             # Dashboard widgets + chart theme
  dates/                 # Date display components
  gifts/                 # Gift form, list, detail
  journal/               # Journal form, timeline, search
  me/                    # User profile section
  people/                # People table, detail sections, avatar uploader, quick actions
  reminders/             # Reminder form, table
  settings/              # Two-panel layout (sidebar nav + detail panel); General settings for date/time format + timezone
lib/                     # Utilities (api-client.ts, auth-context.tsx, query-client.ts, format-datetime.ts, etc.)
routes/                  # TanStack Router file-based routes (_authed/, public/)
schemas/                 # Zod-style TS type definitions (maintained by hand, not generated)
styles.css               # Tailwind + design tokens (:root variables)
```

### Design System (Linear/Stripe Minimal)
- **Accent**: Indigo-600 (#4f46e5) for primary actions, links, and chart data
- **Surfaces**: Zinc palette — white (#ffffff), muted (#fafafa), borders (#e4e4e7)
- **Typography**: Inter (primary), JetBrains Mono (numerics), font-weight 600 (headings)
- **Borders**: Hairline (1px) zinc-200; no shadows (minimal aesthetic)
- **Radius**: 0.375rem (compact aesthetic)
- **Layout**: Horizontal topbar (h-14, sticky, border-b) replaces desktop sidebar
- **Navigation**: TanStack Router with responsive mobile hamburger menu
- **Icons**: Lucide React only; no emojis

### Key Features
- **File-based Routing**: TanStack Router auto-loads `routes/**/*.tsx`; layout nesting via `_layout.tsx`
- **Auth Context**: `lib/auth-context.tsx` stores session & redirects to login if unauthenticated
- **Query Client**: `lib/query-client.ts` configures 5-minute stale time, 10-minute cache duration
- **API Layer**: `endpoints/*.ts` exports fetch functions (POST/GET/PUT/DELETE); handles `X-Requested-With: kith-spa` CSRF header
- **Validation**: Zod schemas in `schemas/` directory (hand-maintained, not generated)
- **Responsive**: Mobile-first; media breakpoints via Tailwind (sm/md/lg/xl)
- **Date/Time Formatting**: `lib/format-datetime.ts` utility converts UTC backend times to user's locale/timezone; settings stored in localStorage (date format, timezone)

### Dashboard Feature
- **Components**: summary-cards, relationship-pulse-chart, action-queue, recent-relationship-activity, upcoming-moments
- **Data Adapter**: `dashboard-data.ts` derives KPIs and shapes API responses for widgets
- **Per-card Refresh**: TanStack Query invalidation on refresh button click
- **Chart Library**: Recharts v3.8.1 with custom indigo/zinc theme

**Validation:**
- Build: `pnpm --dir web build` ✅
- Tests: `pnpm --dir web test` ✅
- Check: `pnpm --dir web check` ✅ (Biome lint/format)
- TypeScript: Clean, no errors
- Responsive: Verified 1280px→mobile

### Authentication Bridge
- **Cookie**: `kith_session` (HttpOnly, SameSite=Lax, Secure when `BEHIND_TLS=true`)
- **CSRF Protection**: All POST/PUT/PATCH/DELETE require `X-Requested-With: kith-spa` header when authenticated by cookie
- **Token Auth**: Server-side only (`TOKEN_AUTH` env var); for future API clients, not SPA

## Key Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `labstack/echo/v5` | v5.1.1+ | HTTP framework |
| `urfave/cli/v3` | v3.8.0+ | CLI subcommands |
| `modernc.org/sqlite` | v1.50.0+ | Pure Go SQLite (no CGO) |
| `golang.org/x/crypto` | latest | bcrypt password hashing |
| `knadh/koanf/v2` | v2.3.4+ | Layered config loading |
| `getsentry/sentry-go` | v0.46.1+ | Error monitoring (optional) |
| `samber/slog-multi` | v1.8.0+ | slog fan-out to multiple handlers |
| `go.uber.org/automaxprocs` | v1.6.0+ | Auto GOMAXPROCS in containers |

## Module & Build

- **Module**: `github.com/nhymxu/kith-pms` — Go 1.26.2+
- **Build**: `make web` (pnpm build → copy SPA) then `CGO_ENABLED=0 go build` for single static binary
- **Binary name**: `kith-pms` (compiled to `bin/kith-pms`)
- **Frontend**: `web/` pnpm workspace; `pnpm build` outputs to `web/dist/`; copied to `internal/web/spa/public/` for embedding

## Test Coverage

15 test files across:
- `auth`: password hashing, session tokens, CSRF token generation
- `audit`: logging behavior, list queries, actor attribution
- `people`: CRUD, search, label associations
- `labels`: CRUD, many-to-many associations
- `journal`: CRUD, FTS5 full-text search
- `dates`: important dates, OnThisDay queries, recurring logic
- `files`: avatar upload, MIME validation, path traversal prevention
- `reminders`: CRUD, completion tracking, status filtering
- `gifts`: CRUD, image operations, debt tracking
- `relationships`: paired rows, self-loop guards, FK constraints

**Total**: 159 Go tests passing. Run all: `make tests` (includes race detector)

React frontend tests: Vitest + @testing-library/react; run via `pnpm --dir web test`
