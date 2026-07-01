# System Architecture

## High-Level Topology

```
┌─────────────────────────────────────────────────────────────────┐
│  Single Binary: kith-pms (CGO_ENABLED=0, no runtime deps)       │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ CLI Layer (urfave/cli v3)                                │   │
│  │  ├─ serve            → starts HTTP server                │   │
│  │  ├─ migrate [up|down] → applies/rolls back migrations    │   │
│  │  ├─ set-password     → interactive password setup        │   │
│  │  ├─ backup --to      → SQLite VACUUM INTO               │   │
│  │  └─ restore --from   → replace database                 │   │
│  └──────────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ Web Layer (Echo v5.2.1 JSON API + React 19 SPA)          │   │
│  │  ├─ /health        → Liveness probe (no auth)            │   │
│  │  ├─ /swagger/*     → Swagger UI + OpenAPI spec (no auth) │   │
│  │  ├─ /v1/*          → JSON REST API (SessionOrBearer)     │   │
│  │  ├─ /assets/*      → Embedded SPA hashed assets (1yr)    │   │
│  │  ├─ /favicon.*     → Embedded favicon (1yr)              │   │
│  │  └─ /*             → index.html catch-all (SPA shell)    │   │
│  │                                                          │   │
│  │  Frontend (React 19 SPA, TanStack Router v1 file-based): │   │
│  │  ├─ /              → Dashboard (KPIs, charts, activity)  │   │
│  │  ├─ /people/*      → People CRUD (table, detail form)    │   │
│  │  ├─ /journal/*     → Journal (list, detail, search)      │   │
│  │  ├─ /gifts/*       → Gifts (form, list, detail)          │   │
│  │  ├─ /reminders/*   → Reminders (form, table)             │   │
│  │  ├─ /dates         → Important dates (upcoming)          │   │
│  │  ├─ /audit         → Audit log (list view)               │   │
│  │  └─ /settings/*    → Labels, rel types, security (PWD)   │   │
│  │                                                          │   │
│  │  API Routes (all /v1, SessionOrBearer auth):              │   │
│  │  ├─ POST /auth/login, logout, logout-all, /me, /password│   │
│  │  ├─ GET/POST/PUT/DELETE /people(/avatar, /relationships)│   │
│  │  ├─ GET/POST/PUT/DELETE /journal, /gifts, /reminders    │   │
│  │  ├─ GET /dates/upcoming, /labels, /relationship-types   │   │
│  │  └─ GET /audit, /me/setup                               │   │
│  └──────────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ Service Layer (auth, people, labels, journal_labels,     │   │
│  │                journal, dates, reminders, gifts, files,   │   │
│  │                audit)                                      │   │
│  │  ├─ Business logic (CRUD, search, validation)            │   │
│  │  └─ Repository patterns (data access abstraction)        │   │
│  └──────────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ Data Layer (SQLite + FTS5)                               │   │
│  │  ├─ WAL mode       → concurrent readers, single writer   │   │
│  │  ├─ Foreign keys   → enforced referential integrity      │   │
│  │  └─ FTS5 triggers  → auto-sync full-text search index    │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
     │
     └─→ SQLite database file (data/kith.db by default)
         ├─ data/kith.db      (main database)
         ├─ data/kith.db-wal  (Write-Ahead Log)
         └─ data/kith.db-shm  (shared memory)
```

## CLI & Dependency Initialization

Entry: `main.go` → `urfave/cli.NewApp().Run(os.Args)`

All subcommands inherit from root command's `Before` hook:
1. **Config loading**: `config.Load()` — three-layer merge via nhymxu/gommon/cfgloader (defaults → .env → env vars)
2. **Logging**: Initialize slog with either text (DEBUG=true) or JSON (DEBUG=false) format; colored output via tint package when DEBUG=true
3. **Sentry (optional)**: If `SENTRY.DSN` set, integrate slog with Sentry error reporting

Subcommands after dependency init:
| Command | Purpose |
|---------|---------|
| `serve` | Start HTTP server on port from `HOST:PORT` (default :8000) |
| `migrate` | Schema management: `migrate up` (apply pending), `migrate down` (rollback latest) |
| `set-password` | Interactive password setup/change (stores Argon2id hash in users table) |
| `backup --to PATH` | SQLite VACUUM INTO PATH; safe to run while server is running |
| `restore --from PATH` | Replace live database with backup; refuses if server modified DB in last 30s |

## Configuration & Environment

Three-layer merge via `nhymxu/gommon/cfgloader` (lowest → highest precedence):

```
1. Hardcoded defaults (configDefaults)  ← baseline
2. .env file (optional, skipped if missing)
3. Environment variables               ← always wins
```

Result unmarshals to global `config.C` (Config struct). Load via `config.Load()` function.

### Supported Environment Variables

| Variable              | Type     | Default          | Purpose                                                 |
|-----------------------|----------|------------------|---------------------------------------------------------|
| `DB_PATH`             | string   | `data/kith.db`   | SQLite database file path                               |
| `DB_AUTO_MIGRATE`     | bool     | `true`           | Apply pending migrations on server startup              |
| `DB_MAX_OPEN_CONNS`   | int      | `1`              | Go connection pool size. Default serialises all writes (prevents `SQLITE_BUSY`). Raise to 5–10 in WAL mode if you also add `PRAGMA busy_timeout`. See *SQLite concurrency* note below. |
| `SESSION_SECRET`      | string   | *(required)*     | Cookie signing secret (min 32 bytes)                    |
| `SESSION_LIFETIME`    | duration | `720h` (30 days) | Session cookie expiry duration                          |
| `BEHIND_TLS`          | bool     | `false`          | Set `true` when behind TLS proxy (marks cookies Secure) |
| `DEBUG`               | bool     | `false`          | `true` → text logs + debug level                        |
| `SENTRY.DSN`          | string   | *(empty)*        | Sentry DSN; omit to disable error reporting             |
| `AVATAR_STORAGE_PATH` | string   | `data/avatars`   | Directory for storing avatar files                      |
| `GIFT_STORAGE_PATH`   | string   | `data/gifts`     | Directory for storing gift images                       |
| `TOKEN_AUTH`          | string   | *(empty)*        | Static bearer token for API clients (optional)          |

### SQLite concurrency note

SQLite's single-writer model means at most one goroutine can hold a write transaction at a time.

**Default (`DB_MAX_OPEN_CONNS=1`):** the Go pool serialises everything at the driver level — no `SQLITE_BUSY` errors, but no concurrent reads either. Code must never hold a transaction open and then make a second query on the same pool (causes a pool deadlock). All service methods that use transactions must prefetch any read dependencies before calling `BeginTx`.

**Raising the limit:** safe to do in WAL mode (already active), which allows concurrent readers alongside one writer. When `DB_MAX_OPEN_CONNS > 1`, `sqlite.go` automatically applies `PRAGMA busy_timeout=5000` so SQLite retries for up to 5 s on write contention instead of returning `SQLITE_BUSY` immediately — no extra configuration needed.

## Logging

- **Handler**: `slog.NewTextHandler` (DEBUG=true) or `slog.NewJSONHandler` (DEBUG=false), both to stdout
- **Fanout**: If Sentry enabled, slogmulti.Fanout writes to base handler + slogsentry (Error level only)
- **All code**: Uses stdlib `log/slog` directly (no third-party logging imports in business logic)

Sentry receives: stack traces (AttachStacktrace: true), all slog Error/above events.

## Prometheus Metrics

**Package**: `internal/metrics/metrics.go`

Exposed at `GET /metrics` (no authentication required) in Prometheus exposition format.

### Metrics Provided

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `http_requests_total` | Counter | method, route, status | Total HTTP request count |
| `http_request_duration_seconds` | Histogram | method, route, status | Request latency buckets |
| `kith_db_size_bytes` | Gauge | - | SQLite database file size (via PRAGMA page_count) |
| `kith_active_sessions` | Gauge | - | Current valid session count |
| `kith_build_info` | Gauge | version, go_version | Build metadata (always 1) |

### Architecture

- Custom Prometheus registry (not global default) to avoid conflicts with imported packages
- HTTP middleware wraps every request: records start time, increments counter on response, observes duration
- Route template normalization strips dynamic parameters (`/people/:id` instead of `/people/42`) to bound label cardinality
- GaugeFunc collectors poll `PRAGMA page_count` and session count on each `/metrics` scrape
- Build info gauge is set once at startup from `cmd/doc.go` version constant

### Testing

- `metrics_test.go` validates: route-template label cardinality is bounded, unknown routes return 404, output format is valid Prometheus exposition.

## Frontend Architecture (React 19 SPA)

### Build Pipeline & Embedding

```
web/src/ (React + TypeScript)
  ├─ pnpm build (Vite 8)
  └─ web/dist/ (hashed assets: index.html, js/*, css/*)
      └─ make web
         └─ copy to internal/api/spa/public/
            └─ go build (//go:embed all:public)
               └─ static binary with SPA embedded
```

**Build Details**:
- **Vite 8**: Fast bundler with code splitting, lazy loading, HMR
- **Output**: `web/dist/` with hashed filenames for cache busting
- **Embedding**: `internal/api/spa/spa.go` uses `//go:embed all:public`
- **Asset Serving**: `/assets/*` served with 1-year immutable cache; `/` serves `index.html` (no-cache) for SPA shell

### Routing & Layouts (TanStack Router v1)

**File-based routing** in `web/src/routes/`:
```
routes/
├── __root.tsx           # Root layout + outlet
├── login.tsx            # Public login page
└── _authed.tsx          # Auth-guarded layout with topbar
    ├── index.tsx        # Dashboard (/)
    ├── network/         # Relationship Network Graph (global view)
    ├── people/          # People CRUD (index, new, $personId, $personId.edit)
    ├── journal/         # Journal (index, new, $entryId, $entryId.edit)
    ├── gifts/           # Gifts (index, new, $giftId, $giftId.edit)
    ├── reminders/       # Reminders (index, new, $reminderId, $reminderId.edit)
    ├── dates/           # Important dates list
    ├── audit/           # Audit log view
    ├── me/              # User profile (index, setup)
    └── settings/        # Settings hub (index, labels, relationship-types, security)
```

**Layout Pattern**: `_authed.tsx` acts as auth guard with shared topbar; all routes under it require authentication.

### Components & Styling

**Component Library**: Shared UI primitives in `web/src/components/ui` use `@base-ui/react` for accessible behavior where primitives are needed, with shadcn-style component APIs restyled for Linear/Stripe minimal:
- **Accent**: Indigo-600 (#4f46e5) — primary actions, links, and interactive elements
- **Surfaces**: Zinc palette (white background, #fafafa muted, #e4e4e7 borders)
- **Borders**: Hairline (1px) zinc-200 throughout; no box shadows
- **Radius**: 0.375rem (compact aesthetic)
- **Typography**: Inter primary, JetBrains Mono for numerics; font-weight 600 headings

**Navigation**: `topbar.tsx` sticky header (h-14, border-b):
- Desktop (md+): Full nav inline with indigo underline active state
- Mobile (<md): Hamburger menu toggle; nav items in collapsible sidebar

**Charts**: Recharts v3.8.1 with custom indigo/zinc theme for dashboard visualizations

**Dashboard Architecture**: The dashboard page (`web/src/routes/_authed/index.tsx`) executes 7 parallel TanStack Query fetches (people, journal, reminders, dates, gifts, labels, audit) to power 5 interactive widgets:
- **Summary KPI Cards**: People count, follow-ups, important dates, gifts, journal activity — per-card refresh via `queryClient.invalidateQueries()`
- **Relationship Pulse Chart**: Recharts line chart with responsive container and custom indigo tooltip; data derived by `dashboard-data.ts`
- **Action Queue**: Filterable pills (all/dates/journal/gifts), capped to 8 rows with Show more/less toggle
- **Recent Relationship Activity**: Capped to 6 entries with timestamps and person links
- **Upcoming Moments**: Capped to 5 upcoming important dates/reminders

**Settings Two-Panel Layout**: `web/src/routes/_authed/settings/` uses a left sidebar with navigation (General, Labels, Relationship Types, Security) and a right detail panel. The General settings panel manages date/time format, timezone, and audit log retention days. Settings persist to the `user_setting` table and are loaded on app start.

**People Detail Page (9 Sections)**: The people detail view at `web/src/features/people/detail/sections/` is modularized into 9 focused section components orchestrated by `person-detail-sections.tsx`:
1. **Overview-section.tsx** — Inline edit mode toggle with avatar controls
2. **Contacts-section.tsx** — Inline add/edit/delete per row
3. **Locations-section.tsx** — Inline add/edit/delete per row
4. **Labels-section.tsx** — Attached/available sub-headings for clear distinction
5. **Relationships-section.tsx** — Notes truncated below entries; search-as-you-type person picker
6. **Journal-section.tsx** — "Quick journal" button; entry rows show other people involved
7. **Work-history-section.tsx** — Inline add/edit/delete via bulk replace API
8. **Important-dates-section.tsx** — DOB read-only with lock icon; custom dates add/edit/delete
9. **Gifts-section.tsx** — "Quick gift" button in section header
Each section uses `SectionCard` for consistent visual spacing and TanStack Query optimistic updates.

**Relationship Network Visualization**: The network page (`web/src/routes/_authed/network.tsx`) displays a force-directed graph of your entire contact network using `react-force-graph-2d` v1.29.1. Features:
- **Global View**: Full network showing all people (nodes) with relationships (edges); click any person to navigate to detail view
- **Per-Person Ego View**: Accessible from people detail page; shows selected person + direct connections (1-hop); helps visualize local community
- **Visual Encoding**: Nodes display person avatar (cached from API); edges colored by relationship type; optional group-by-label coloring (toggle in UI)
- **Interaction**: Drag nodes to pin in position; zoom/pan; recenter button to reset view; canvas resized to container
- **Graph Data**: `web/src/lib/graph-data.ts` utility processes people + relationships into force-graph node/link format with avatar URL caching; lazy-loaded on route entry (separate code chunk)

### Data Layer (TanStack Query v5)

**Configuration**: 5-minute stale time, 10-minute cache duration
**Endpoints**: API functions in `web/src/endpoints/*.ts` per resource:
- `audit.ts`, `auth.ts`, `dates.ts`, `gifts.ts`, `journal.ts`, `labels.ts`, `me.ts`, `people.ts`, `relationship-types.ts`, `reminders.ts`

**API Client**: `lib/api-client.ts` shared fetch wrapper:
- Attaches `X-Requested-With: kith-spa` CSRF header for POST/PUT/PATCH/DELETE
- Handles cookie-based auth automatically
- Supports Bearer token auth (for future programmatic access)

**Auth Context**: `lib/auth-context.tsx` manages session state:
- Stores user info in React context
- Redirects to login if unauthenticated
- Provides `useAuth()` hook for component consumption

### Validation & Forms (TanStack Form v0 + Zod)

**Schema Location**: `web/src/schemas/*.ts` (hand-maintained, not generated):
- `audit.ts`, `auth.ts`, `date.ts`, `gift.ts`, `journal.ts`, `label.ts`, `person.ts`, `relationship-type.ts`, `reminder.ts`
- Must exactly match Go API types: case-sensitive field names, optional fields, enum values

**Form Pattern** (example):
```typescript
// schemas/person.ts
export const PersonSchema = z.object({
  id: z.number(),
  name: z.string(),
  dateOfBirth: z.date().nullable(),
  labels: z.array(z.object({ id: z.number(), name: z.string() }))
});

// Component
export function PersonForm() {
  const form = useForm({ defaultValues, onSubmit: (values) => savePerson(values) });
  return <form onSubmit={form.handleSubmit}>...</form>;
}
```

---

## HTTP Server (Echo v5.2.1)

### Handler Architecture

**Handler Package Pattern** (`internal/api/handler/` subdirectory):
- All HTTP handlers organized in a dedicated `handler/` subpackage (moved from flat `internal/api/handlers_*.go` files)
- Struct-based handlers with injected service dependencies
- Pattern: `type XxxAPI struct { Svc *xxx.Service }` with method receivers `(h *XxxAPI) MethodName(c echo.Context) error`
- Centralized response helpers in `response.go`: `ok(c, data)`, `created(c, data)`, `apiErr(c, code, msg)` with {data, error} envelope
- 20+ handler files organized by domain (auth, people, labels, journal, dates, gifts, reminders, audit, relationships, work_history, avatars, people_labels, people_quick, me) plus testhelpers

### Global Middleware Stack (applied in order)

| Middleware | Purpose |
|-----------|---------|
| `middleware.Recover()` | Catches panics, returns 500 |
| `middleware.RequestID()` | Attaches unique request ID header to each request |
| `middleware.Gzip()` | Response compression (deflate/gzip) |
| `middleware.RequestLogger()` | Structured access logging via slog |
| Sentry middleware | Auto-captures request errors → Sentry |

### Routes (mounted in `internal/api/mount.go`)

```
/health                → GET (liveness probe, no auth)
/ready                 → GET (readiness probe: DB pingable + migrations applied, no auth)
/metrics               → GET (Prometheus metrics, no auth)
/swagger/*             → Swagger UI + OpenAPI spec (no auth)
/v1/*                  → JSON REST API (see api package)
/assets/*              → Embedded SPA hashed assets (1-year cache, immutable)
/favicon.*             → Embedded favicon (1-year cache)
/*                     → index.html catch-all (200, no-cache, CSP headers) — SPA handles all routing
```

#### JSON API Routes (`/v1/*`)

**Auth** (SessionOrBearer, rate-limited):
- `POST /auth/login` — create session (5 attempts per 15 min)
- `POST /auth/logout` — destroy current session
- `POST /auth/logout-all` — destroy all sessions for user
- `GET /auth/me` — current user or 401

**User Profile**:
- `GET /me` — self profile person
- `POST /me/setup` — designate self person
- `POST /auth/password` — change password (requires current)

**People** (CRUD + relationships):
- `GET /people`, `POST /people` (list/search + create)
- `GET/PUT/DELETE /people/:id` (detail, update, delete)
- `POST/GET/DELETE /people/:id/avatar` (upload, retrieve, delete)
- `GET /people/:id/labels` — person's labels
- `GET /people/:id/relationships` — person's relationships
- `POST/DELETE /people/:id/relationships` — attach/detach relationships

**Journal**, **Reminders**, **Dates** — similar CRUD patterns

**Gifts** (CRUD + image storage):
- `GET /gifts`, `POST /gifts` (list + create)
- `GET/PUT/DELETE /gifts/:id` (detail, update, delete)
- `POST/GET/DELETE /gifts/:id/image` (upload, retrieve, delete image; stored in `GIFT_STORAGE_PATH`)
- `GET /people/:id/gifts` — person's gifts

**Labels** — CRUD

**Relationships** — GET types, POST/DELETE type defs

**Audit** — GET audit log entries, `POST /audit/cleanup` (manual purge of entries older than retention period)

**Settings** — `GET /settings`, `PUT /settings` (user preferences including audit log retention days)

All state-changing calls (POST/PUT/PATCH/DELETE) require `X-Requested-With: kith-spa` header when authenticated by cookie.

### OpenAPI/Swagger Documentation

**Endpoint**: `/swagger/index.html` — interactive Swagger UI (no authentication required)

**Generation**:
- Annotations: All 16 handler files in `internal/api/handler/` include swaggo v2 comments documenting request/response schemas, parameters, and security
- Package-level annotations: `cmd/doc.go` defines API title, version, base path (`/v1`), security schemes (CookieAuth + Bearer token)
- Generation: `make swagger` runs `swag init -g cmd/doc.go -o docs` to generate OpenAPI 2.0 spec
- Integration: Build pipeline includes `make swagger` before final binary compilation

**Spec Files** (auto-generated, committed to repo):
- `internal/api/swagger/swagger.json` — OpenAPI 2.0 spec in JSON format
- `internal/api/swagger/swagger.yaml` — OpenAPI 2.0 spec in YAML format
- `internal/api/swagger/docs.go` — Go package with embedded spec

**Custom Handler**: `internal/api/swagger_handler.go` provides Echo v5-compatible handler wrapping `swaggo/files/v2` embedded assets; no external `labstack/echo-swagger` v4 dependency.

### Session & Auth Flow

```
1. SPA calls POST /v1/auth/login with JSON {password}
   ↓
2. Handler validates password (Argon2id verify vs users table)
   ↓
3. Create session token:
   - Generate cryptographic session ID
   - Store session in database (user_sessions table, expires_at)
   - Sign session ID + user ID → HMAC-SHA256 token
   ↓
4. Set session cookie (HttpOnly, SameSite=Lax; Secure when BEHIND_TLS=true)
   - Name: kith_session
   - Value: HMAC-signed token
   - Path: /
   - MaxAge: SESSION_LIFETIME (default 720h = 30 days)
   ↓
5. On subsequent /v1/* requests:
   - SessionOrBearer middleware: accepts either cookie or Authorization: Bearer token
   - **Cookie path**: validates HMAC, looks up session, verifies expiry, injects user context
   - **Bearer path**: validates static TOKEN_AUTH value (future machine clients only)
   - **CSRF check**: All state-changing calls (POST/PUT/PATCH/DELETE) require X-Requested-With: kith-spa
     header when authenticated by cookie (protects against cross-site form submissions)
   ↓
6. Password change (POST /v1/auth/password):
   - Verify current password
   - Hash new password with Argon2id (golang.org/x/crypto/argon2)
   - Invalidate all sessions for that user
   - Rate limited: 5 attempts per 15 minutes
   ↓
7. Logout (POST /v1/auth/logout):
   - Clear session record in database
   - Clear session cookie in response
```

## File Storage Layer

### Avatar Storage Architecture

**Location**: `internal/files/service.go`

The file storage layer handles avatar uploads with security and durability guarantees:

**FileService Interface (Avatars & Documents)**:
- `SaveAvatar(personID, file, header)` → saves file (multipart), returns relative path (5MB limit, MIME allowlist)
- `SaveAvatarBytes(personID, data, mimeType)` → saves raw bytes for imports (5MB limit, MIME allowlist)
- `SaveDocument(personID, data, originalName)` → saves document (any MIME type, 50MB limit, no allowlist)
- `DeleteAvatar(personID, path)` → removes file and cleans up empty directories
- `GetAvatarPath(personID)` → returns base directory for person's avatars

**LocalFileService Implementation**:
- **Avatar base directory**: Configured via `AVATAR_STORAGE_PATH` (default: `data/avatars`)
- **Avatar structure**: `data/avatars/{personID}/{randomStr}-{sanitized-name}.{ext}`
- **Document base directory**: Uses same base `data/` with `documents/` subdirectory
- **Document structure**: `data/avatars/documents/{personID}/{randomStr}-{sanitized-name}.{ext}` (stored under avatars base for simplicity)
- **Atomic writes**: Temp file → sync → rename (prevents partial uploads)
- **Path traversal prevention**: Validates clean path stays within base directory

**Security Controls (Avatars)**:
- **MIME validation**: Dual-check (HTTP header + magic number via `http.DetectContentType`)
- **Allowed types**: `image/jpeg`, `image/png`, `image/gif`, `image/webp` only
- **Size limit**: 5MB per file (enforced at handler + service layer)
- **Filename sanitization**: Alphanumeric + dash/underscore only; max 50 chars
- **Random prefix**: 8-byte hex prefix prevents filename collisions and guessing

**Security Controls (Documents)**:
- **No MIME allowlist**: Accepts any file type (since documents come from trusted Monica exports)
- **Size limit**: 50MB per file (much higher than avatars for diverse document types)
- **Filename sanitization**: Preserves original extension; removes special chars from base name
- **Random prefix**: 8-byte hex prefix for uniqueness

**Integration with People Service (Avatars)**:
- Service stores avatar metadata in database: `avatar_path`, `avatar_mime_type`, `avatar_size`, `avatar_uploaded_at`
- On upload: saves file first, then updates DB in transaction; rolls back file on DB error
- On delete: clears DB metadata, then removes file (best-effort cleanup)
- On replace: saves new file, updates DB, then deletes old file (old file survives DB errors)

**Integration with Monica Import (Documents)**:
- During `monica-import`, documents extracted from Monica export are saved via `SaveDocument()`
- Each document becomes a DOCUMENT-labelled journal entry linked to the contact
- No database metadata beyond the journal entry itself; documents stored for archival/reference

### Gift Image Storage Architecture

**Location**: `internal/api/handler/gifts.go` (image endpoints)

Gift images are stored separately from avatars with similar security patterns:

**Storage Configuration**:
- **Base directory**: Configured via `GIFT_STORAGE_PATH` (default: `data/gifts`)
- **File naming**: Gift ID as filename (e.g., `123.jpg`); MIME type stored in database

**Image Endpoints**:
- `POST /v1/gifts/:id/image` — upload image (multipart form, max 5MB)
- `GET /v1/gifts/:id/image` — retrieve image (served with 24-hour cache header)
- `DELETE /v1/gifts/:id/image` — remove image and clear metadata

**Security Controls**:
- **MIME validation**: Magic number detection via `http.DetectContentType`
- **Allowed types**: `image/jpeg`, `image/png`, `image/gif`, `image/webp`
- **Size limit**: 5MB per file
- **Path traversal prevention**: Validates clean path stays within `GIFT_STORAGE_PATH`

**Integration with Gifts Service**:
- Service stores image metadata in database: `image_path`, `image_mime_type`
- On upload: validates file, stores in `GIFT_STORAGE_PATH`, updates DB metadata
- On delete: clears DB metadata, removes file from disk
- On retrieve: serves from disk with cache headers

**MIME Type Detection**: File type is detected from file extension at serve time, not stored in database. This simplifies schema and avoids MIME type spoofing attacks.

## Database Layer & ORM

### SQLite Configuration

**File**: `internal/db/sqlite.go`

Connection settings:
- **Driver**: `modernc.org/sqlite` (pure Go, no CGO)
- **ORM Wrapper**: uptrace/bun (query builder + struct scanning)
- **Raw SQL Approach**: All queries are raw SQL strings; bun used as thin wrapper for execution and struct scanning
- **No ORM Models**: Intentionally skipped bun model layer; kept raw SQL throughout for simplicity
- **WAL mode**: Write-Ahead Log for concurrent readers without blocking writer
- **Foreign keys**: Enabled (PRAGMA foreign_keys=ON)
- **Synchronous**: NORMAL (safe with WAL; balance of durability vs speed)
- **MaxOpenConns**: 1 (serializes all writes per SQLite single-writer model)

### Query Execution Pattern

All 11 repositories accept `bun.IDB` interface (satisfied by `*bun.DB` or `bun.Tx`) for transaction support:
```go
func (r *Repo) GetByID(ctx context.Context, db bun.IDB, id int64) (*Model, error) {
    // Raw SQL execution via bun
    err := db.NewRaw("SELECT * FROM table WHERE id = ?", id).Scan(ctx, &result)
    return result, err
}
```

**FTS5 Queries**: Journal full-text search uses `db.NewRaw()` pattern with direct SQL:
```go
// FTS5 virtual table queries
err := db.NewRaw("SELECT activities.* FROM activities WHERE rowid IN (SELECT rowid FROM activities_fts WHERE activities_fts MATCH ?)", term).Scan(ctx, &results)
```

**Debug Logging**: When `DEBUG=true`, `bundebug.NewQueryHook()` is conditionally registered for verbose query logging to stderr.

### Schema & Migrations

**Location**: `internal/db/migrations/`

| Migration | Purpose |
|-----------|---------|
| `0001_init.sql` | users, people, contacts, locations, labels, label_assignments tables |
| `0002_user_session.sql` | user_sessions table (session_id, session_token, expires_at) |
| `0003_person.sql` | refine person table structure |
| `0004_label.sql` | refine label-person association |
| `0005_activity.sql` | journal entries (activities) + links to people |
| `0006_activity_fts.sql` | FTS5 virtual table + triggers for full-text search |
| `0007_important_date.sql` | important_date table with virtual month_day column for date queries |
| `0008_reminder.sql` | reminders table with person/date associations and completion tracking |
| `0009_person_avatar.sql` | avatar_path, avatar_mime_type, avatar_size, avatar_uploaded_at columns on person table |
| `0010_work_history.sql` | work history table |
| `0011_audit_log.sql` | audit_log table for entity change tracking (entity_type, entity_id, entity_name, action, actor_id, created_at) |
| `0012_gift.sql` | gift table with direction (gave/received), debt_type (owed/owe), person association, and image storage metadata |
| `0013_person_self.sql` | is_self column on person table with unique index for self-profile feature |
| `0014_person_last_contact.sql` | last_contact_at nullable timestamp column on person table with partial index |
| `0015_relationship_type.sql` | relationship_type table (name, reverse_name, self-FK inverse_type_id) + person_relationship junction table (from_person_id, to_person_id, relationship_type_id, notes) with UNIQUE and CHECK constraints |
| `0016_user_setting.sql` | user_setting table (key, value) for storing user preferences (date_format, time_format, timezone, audit_log_retention_days) |
| `0017_reminder_recurrence.sql` | recurrence_rule TEXT and recurrence_end_date TEXT columns on reminder table for storing recurrence configuration |
| `0018_person_gender.sql` | gender TEXT column on person table |
| `0019_rename_people_labels.sql` | rename label table to people_labels for clarity (distinct from journal_label) |
| `0020_journal_label.sql` | journal-specific labels (separate from people labels) with color support |
| `0021_person_nickname_lower.sql` | generated lowercase nickname column for case-insensitive search |
| `0022_drop_mime_type_columns.sql` | remove avatar_mime_type and gift_image_mime_type; MIME now detected at serve-time |
| `0023_work_history.sql` | work history table with employment dates and employment records |
| `0024_audit_metadata.sql` | add metadata TEXT column to audit_log for structured field-level change tracking |

**Total Migrations**: 24 migrations applied in order; tracked via schema_migrations table.

**Loading**: `internal/db/migrations.go` — loads SQL files in order, tracks applied versions in schema_migrations table.

### Reminder Recurrence, Birthday Reminders, & Settings Notes

**Reminders**: Support 8 recurrence types (daily, weekly, monthly, yearly, custom interval, day-of-week, relative-to-last-contact, **birthday**) with optional end date cutoff. Auto-spawn next occurrence on completion via pure `computeNextDue()` function.

**Birthday Reminders** (new in Phase 4.5):
- Triggered by `recurrence_type: "birthday"` in RecurrenceRule JSON
- Anchored to `person.date_of_birth` field with annual recurrence
- Configurable `days_before_dob` field (0–30 day advance warning presets)
- DOB sync logic: on DOB update, re-computes pending birthday reminder due dates; on DOB clear, deletes all associated birthday reminders
- No new DB columns required — stored entirely in existing `recurrence_rule` JSON
- UI: person-form checkbox "Create annual birthday reminder" (DOB-conditional); reminder-form toggle with picker; mutual exclusion with recurring checkbox for simplicity
- Handles edge cases: Feb 29 birthdays, yearless dates, leap year transitions

**Settings Persistence**: Key-value settings stored in `user_setting` table (server-side) with matching localStorage keys in frontend for date format, timezone, audit log retention policy, and other user preferences. Ensures settings persist across sessions and are loaded on app initialization.

### FTS5 Full-Text Search

**Architecture**:
- Virtual FTS5 table: `activities_fts` (mirrors activities.title + content)
- Triggers: Auto-update FTS5 on INSERT/UPDATE/DELETE of activities
- Query: Search via `activities_fts.rowid` with MATCH clause

**Example search**:
```sql
SELECT activities.* FROM activities
WHERE rowid IN (SELECT rowid FROM activities_fts WHERE activities_fts MATCH 'search term')
```

### Audit Logging & Retention

**Architecture**:
- Service: `internal/audit/Service` — logs all entity mutations (CREATE, UPDATE, DELETE)
- Integration: Injected into people, journal, labels, reminders, dates, work_history, gifts, relationships services
- Best-effort: Logging failures never block primary operations; errors logged as warnings only
- Actor attribution: `audit.WithActor(ctx, userID)` and `ActorFromCtx(ctx)` for context-based user tracking
- Storage: `audit_log` table (entity_type, entity_id, entity_name, action, actor_id, created_at, **metadata**)
- **Field-Level Tracking** (new in Phase 4.5): `metadata` column stores structured JSON array of `Change` objects tracking field-level mutations:
  ```
  {
    "changes": [
      {"field": "name", "old_value": "Alice", "new_value": "Alice Smith"},
      {"field": "date_of_birth", "old_value": null, "new_value": "1990-01-15"}
    ]
  }
  ```
- **Retention Policy**: Configurable TTL via `audit_log_retention_days` setting (0 = disabled/keep forever)
- **Purge Method**: `Repo.Purge(ctx, db, days)` deletes entries older than N days using SQLite datetime arithmetic
- **Manual Cleanup**: `POST /v1/audit/cleanup` endpoint triggers immediate purge, returns `{"deleted": N}`

**Usage**:
```go
s.auditSvc.Log(ctx, audit.EntityType("person"), id, "Alice Smith", audit.ActionCreated, metadata)
s.auditSvc.Purge(ctx, retentionDays) // Auto-purge based on setting
```

## Entity Relationships

```
users (1)
  ├─ (1:N) user_sessions
  └─ (1:N) people (created_by)

people (1)
  ├─ (1:N) contacts (phone, email)
  ├─ (1:N) locations (address)
  ├─ (1:N) important_date (birthdays, anniversaries, milestones)
  ├─ (1:N) reminders (optional person association)
  ├─ (1:N) gifts (gift records with direction + debt tracking)
  ├─ (N:M) labels (via label_assignments)
  ├─ (N:M) activities (via activity_links)
  ├─ (N:M) person_relationship (bidirectional relationships)
  └─ (0:1) self profile (is_self flag, unique constraint)

labels (1)
  └─ (N:M) people (via label_assignments)

activities (1)  [Journal entries]
  ├─ (N:M) people (via activity_links)
  └─ (1:1 virtual) activities_fts [FTS5 index]

reminders (1)
  ├─ (N:1) people (optional FK)
  └─ (N:1) important_date (optional FK)

gifts (1)
  └─ (N:1) people (FK to person who gave/received gift)

relationship_type (1)
  ├─ (0:1) self-referential inverse_type_id
  └─ (1:N) person_relationship (FK)

person_relationship (N)
  ├─ (N:1) from_person_id (FK to people)
  ├─ (N:1) to_person_id (FK to people)
  └─ (N:1) relationship_type_id (FK to relationship_type)
```

## Deployment

### Single Binary
- Compiled with `CGO_ENABLED=0` → static binary, no libc or runtime dependencies
- Embedded React SPA (`//go:embed all:public` in `internal/api/spa/spa.go`)
  - Vite builds React 19 SPA to `web/dist/`
  - `make web` copies `web/dist/` → `internal/api/spa/public/`
  - `go embed` compiles all assets into binary (hashed filenames for cache busting)
- Embedded migrations (SQL files compiled into binary)
- **Build pipeline**: `make web` (pnpm build → copy) → `make build` (CGO_ENABLED=0 go build)`
- **Asset serving**: `/assets/*` served with 1-year cache headers; `/` serves index.html with no-cache + CSP headers

### Container (multi-stage Dockerfile)
- Stage 1: `node:24-alpine` — `pnpm install --frozen-lockfile && pnpm build`
- Stage 2: `golang:1.26.4-alpine` — copies SPA into embed path, runs `go build`
- Stage 3: `gcr.io/distroless/static-debian12` — minimal runtime, non-root UID 65532 (also `debian-slim` variant available)
- `go.uber.org/automaxprocs` auto-sets GOMAXPROCS to match container CPU quota
- Database: Mount volume at `/data` for persistent storage

### Docker Compose
- **Development**: `compose.dev.yml` at repository root for local setup
- **Production**: `deploy/compose/docker-compose.yml` with Litestream sidecar for continuous SQLite replication to S3-compatible storage
- **Legacy**: `compose.yml` at repository root (use `compose.dev.yml` for development)
- **Litestream Init Pattern**: Init container restores database from S3 before app starts; replicate sidecar streams WAL frames continuously
- **Multi-platform**: Built and pushed to GitHub Container Registry (GHCR) on release

### Kubernetes Deployment
- **Manifests**: `deploy/k8s/base/` with Kustomize base layer (Namespace, Deployment, Service, PVC, Secret, ConfigMap)
- **Optional Components**: `deploy/k8s/components/` for Ingress (cert-manager), ServiceMonitor (Prometheus Operator)
- **Overlays**: Example overlay at `deploy/k8s/overlays/example/` for flexible composition
- **Security**: Non-root user (65532), read-only root filesystem, dropped capabilities
- **Database**: PVC for persistent SQLite storage; Litestream sidecar for backup to S3
- **Replicas**: Hard-coded to 1 with `Recreate` strategy to prevent SQLite corruption from multi-writer conflicts

### Safety Features
- Automatic migrations on startup (if DB_AUTO_MIGRATE=true)
- Backup via VACUUM INTO (safe while server running)
- Restore with 30-second server-activity heuristic (prevents data loss)
- Health check at `/health` and readiness check at `/ready` for container orchestration
