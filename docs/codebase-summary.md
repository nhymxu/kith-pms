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
│   │   └── migrations.go         # SQL migration loader and executor
│   ├── auth/                     # Authentication & session management
│   │   ├── domain.go             # User, Session, CSRF token data structures
│   │   ├── password.go           # Argon2id hashing & verification
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
│   ├── files/                    # File storage service
│   │   ├── service.go            # LocalFileService for avatar uploads
│   │   └── service_test.go       # File service unit tests
│   ├── monica/                   # Monica PRM import package
│   │   ├── parser.go             # Monica JSON export format unmarshaling
│   │   ├── mapper.go             # Field mapping from Monica to kith-pms domain
│   │   └── mapper_test.go        # Unit tests for Monica-to-domain mapping
│   └── web/                      # HTTP handler layer
│       ├── server.go             # Echo setup (global middleware)
│       ├── route.go              # Route mounting: /health, /v1/*, spa.Handler()
│       ├── spa/                  # Embedded React SPA
│       │   ├── spa.go            # //go:embed all:public; Handler() with catch-all fallback
│       │   └── public/           # Populated by `make web` (gitignored except placeholder.txt)
│       └── forms/                # Form utilities (pure Go, no templ dependency)
│           └── forms.go          # ParseIndexed, IndexedName helpers
├── pkg/                          # Shared packages
│   ├── config/                   # Configuration management
│   │   ├── env.go                # LoadConfig(), EnvConfigMap, environment parsing
│   │   ├── const.go              # Application constants
│   │   └── default.go            # Default config values
│   └── errors/                   # Custom error types (planned)
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
├── docs/                         # Project documentation
├── Dockerfile                    # Multi-stage container build
├── Makefile                      # Dev workflow targets
├── go.mod / go.sum               # Module definition and lockfile
├── templ.iml                     # Templ IDE config (optional)
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
- **password.go**: Argon2id hashing (via golang.org/x/crypto) with verification
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
- **domain.go**: Reminder (title, notes, due_date, person_id, important_date_id, completed), ReminderWithPerson
- **service.go**: CRUD for reminders; completion tracking; filter by status and person
- **repo.go**: Queries for reminders with person joins; status filtering
- **service_test.go**: Integration tests for reminder CRUD and completion

### `internal/gifts` — Gift management & debt tracking
- **domain.go**: Gift (title, description, direction, debt_type, person_id, image_path), GiftWithPerson; Direction and DebtType enums
- **service.go**: CRUD for gifts; UploadImage/DeleteImage for gift photos; persists metadata (path, MIME type, size, upload timestamp)
- **repo.go**: Queries for gifts with person joins; UpdateImage metadata updates
- **service_test.go**: Integration tests for gift CRUD and image operations

### `internal/relationships` — Person-to-person relationships
- **domain.go**: RelationshipType (name, reverse_name, optional inverse_type_id), PersonRelationship (from/to person IDs, type, notes), RelationshipView (rendered relationship with resolved type names)
- **service.go**: CreateType/UpdateType/DeleteType for relationship types; AttachRelationship/DetachRelationship for person junctions; handles symmetric and asymmetric paired types with bidirectional row creation
- **repo.go**: sqlRelationshipTypeRepo and sqlPersonRelationshipRepo; paired transaction writes for bidirectional relationships; FindPair for locating inverse rows
- **service_test.go**: 10 integration tests covering paired rows, self-loop guards, FK constraints, symmetric type bidirectionality

### `internal/files` — File storage service
- **service.go**: LocalFileService for avatar uploads with MIME validation, size limits, path traversal prevention
- **service_test.go**: File service unit tests

### `internal/monica` — Monica PRM data import
- **parser.go**: Unmarshals Monica JSON export format (contacts, activities, reminders, tags, etc.) into typed structs
- **mapper.go**: Pure-function mapping from Monica domain types to kith-pms domain types (Person, ContactInfo, Location, Activity, Reminder, ImportantDate)
- **mapper_test.go**: Unit tests for edge cases (birthdate year handling, contact type classification, name assembly, tag deduplication)

### `internal/web` — HTTP layer
- **server.go**: Creates Echo instance with global middleware (recover, request ID, gzip, logger, sentry)
- **route.go**: Mounts `/health`, API (`/v1/*` via `api.Mount`), then `spa.Handler()` as catch-all; injects session loader + audit actor middleware
- **spa/spa.go**: Embeds `public/` via `//go:embed all:public`; serves `/assets/*` with 1-year immutable cache; returns `index.html` (no-cache, CSP headers) for all non-API GET paths; real 404 for unknown `/assets/*` paths
- **spa/public/**: Populated at build time by `make web` (copies `web/dist/` here); gitignored except `placeholder.txt` sentinel
- **forms/**: Pure Go form parsing utilities; `ParseIndexed` for multi-row form fields

## React SPA Frontend (`web/` directory)

### Dashboard Feature (`web/src/features/dashboard/`)

**Interactive Relationship Dashboard** — Premium responsive PRM dashboard for daily use.

**Components:**
- `dashboard-data.ts` — Pure data adapter; derives KPIs, chart points, action queue, activity feed, upcoming moments from API responses
- `summary-cards.tsx` — KPI cards (people count, follow-ups, dates, gifts, journal activity) with refresh icons
- `relationship-pulse-chart.tsx` — Recharts line chart with responsive container and custom tooltip
- `action-queue.tsx` — Filterable list with pills, capped to 8 rows, Show more/less toggle
- `recent-relationship-activity.tsx` — Recent activity capped to 6 entries
- `upcoming-moments.tsx` — Upcoming moments capped to 5 entries
- `dashboard-card.tsx` — Reusable card primitive with title, subtitle, Lucide icon, refresh action, loading/stale/error slots
- `dashboard-filter-pill.tsx` — Filter pill with active/inactive/hover/focus states
- `empty-state.tsx` — Empty state component for widgets with no data
- `chart-theme.ts` — Chart color palette (indigo-600 primary, zinc surfaces, hairline borders)

**Design System (Linear/Stripe Minimal):**
- **Accent**: Indigo-600 (#4f46e5) for primary actions and chart data
- **Surfaces**: Zinc palette (white, #fafafa muted, #e4e4e7 borders) with no shadows
- **Typography**: Inter primary font, JetBrains Mono for numerics, font-weight 600 for headings
- **Borders**: Hairline (1px) zinc-200 borders; no box shadows
- **Radius**: 0.375rem (compact, minimal rounding)
- **Responsive grid**: 1280px+ desktop (4–5 KPI columns, 8/4 split main), tablet (2 columns), mobile (single column)
- **Icons**: Lucide only; no emojis
- **Hover states**: Subtle background shifts on cards/list rows; active filter pills highlighted with indigo underline

**Navigation:**
- **Topbar** (h-14, sticky, border-b zinc-200): Horizontal navigation bar replacing desktop sidebar
- **Desktop** (md+): Full nav items visible inline with active state underline (indigo)
- **Mobile** (<md): Hamburger menu toggle; nav items in collapsible sidebar
- **NavLink component**: Supports "topbar" variant with underline active state

**Data Flow:**
- Route: `web/src/routes/_authed/index.tsx`
- TanStack Query fetches existing endpoints (people, journal, reminders, dates, gifts, audit, me)
- Dashboard adapter derives metrics and shapes data for widgets
- Per-card refresh invalidates relevant queries only
- Handles empty data, partial failures, and last-known-value states

**Dependencies:**
- `recharts ^3.8.1` — Chart library
- `@tanstack/react-query` — Data fetching and caching
- `lucide-react` — Icon library
- `tailwindcss` — Styling

**Validation:**
- Build: `pnpm --dir web build` ✅
- Tests: `pnpm --dir web test` ✅ (10/10 passing)
- Check: `pnpm --dir web check` ✅
- TypeScript: Clean, no errors
- Biome: No linting issues
- Browser: Responsive 1280px→mobile; all interactive states verified

---
- **env.go**: LoadConfig() with three-layer merge (defaults → .env file → env vars); unmarshals to global ENV
- **const.go**: Application constants (timezones, timeouts, etc.)
- **default.go**: configDefaults map (lowest precedence layer)

## Key Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `labstack/echo/v5` | v5.1.0+ | HTTP framework |
| `urfave/cli/v3` | v3.8.0+ | CLI subcommands |
| `modernc.org/sqlite` | latest | Pure Go SQLite (no CGO) |
| `golang.org/x/crypto` | latest | Argon2id password hashing |
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

10 test files across:
- `auth`: password hashing, session tokens, CSRF token generation
- `audit`: logging behavior, list queries, actor attribution
- `people`: CRUD, search, label associations
- `labels`: CRUD, many-to-many associations
- `journal`: CRUD, FTS5 full-text search
- `dates`: important dates, OnThisDay queries, recurring logic
- `files`: avatar upload, MIME validation, path traversal prevention
- `reminders`: CRUD, completion tracking, status filtering

**Total**: 114 tests passing. Run all: `make tests` (includes race detector)
