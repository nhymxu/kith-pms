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
│  │ Web Layer (Echo v5 JSON API + React SPA)                 │   │
│  │  ├─ /health        → Liveness probe (no auth)            │   │
│  │  ├─ /v1/*          → JSON REST API (SessionOrBearer)     │   │
│  │  ├─ /assets/*      → Embedded SPA hashed assets (1yr)    │   │
│  │  ├─ /favicon.*     → Embedded favicon (1yr)              │   │
│  │  └─ /*             → index.html catch-all (SPA shell)    │   │
│  │                                                          │   │
│  │  Frontend (React 19 SPA, TanStack Router file-based):     │   │
│  │  ├─ /              → Dashboard (KPIs, charts, activity)  │   │
│  │  ├─ /people/*      → People CRUD (table, detail form)    │   │
│  │  ├─ /journal/*     → Journal (list, detail, search)      │   │
│  │  ├─ /gifts/*       → Gifts (form, list)                  │   │
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
│  │ Service Layer (auth, people, labels, journal, dates,     │   │
│  │                reminders, gifts, files, audit)            │   │
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
1. **Config loading**: `config.LoadConfig()` — three-layer merge (defaults → .env → env vars)
2. **Logging**: Initialize slog with either text (DEBUG=true) or JSON (DEBUG=false) format
3. **Sentry (optional)**: If SENTRY_DSN set, integrate slog with Sentry error reporting

Subcommands after dependency init:
| Command | Purpose |
|---------|---------|
| `serve` | Start HTTP server on port from `HOST:PORT` (default :8000) |
| `migrate` | Schema management: `migrate up` (apply pending), `migrate down` (rollback latest) |
| `set-password` | Interactive password setup/change (stores Argon2id hash in users table) |
| `backup --to PATH` | SQLite VACUUM INTO PATH; safe to run while server is running |
| `restore --from PATH` | Replace live database with backup; refuses if server modified DB in last 30s |

## Configuration & Environment

Three-layer merge via koanf (lowest → highest precedence):

```
1. configDefaults (hardcoded)     ← baseline
2. .env file (dotenv)             ← optional, skipped if missing
3. Environment variables          ← always wins
```

Result unmarshals to global `config.ENV` (`EnvConfigMap`).

### Supported Environment Variables

| Variable | Type | Default | Purpose |
|----------|------|---------|---------|
| `DB_PATH` | string | `data/kith.db` | SQLite database file path |
| `DB_AUTO_MIGRATE` | bool | `true` | Apply pending migrations on server startup |
| `SESSION_SECRET` | string | *(required)* | Cookie signing secret (min 32 bytes) |
| `SESSION_LIFETIME` | duration | `720h` (30 days) | Session cookie expiry duration |
| `BEHIND_TLS` | bool | `false` | Set `true` when behind TLS proxy (marks cookies Secure) |
| `DEBUG` | bool | `false` | `true` → text logs + debug level |
| `SENTRY_DSN` | string | *(empty)* | Sentry DSN; omit to disable error reporting |
| `AVATAR_STORAGE_PATH` | string | `data/avatars` | Directory for storing avatar files |

## Logging

- **Handler**: `slog.NewTextHandler` (DEBUG=true) or `slog.NewJSONHandler` (DEBUG=false), both to stdout
- **Fanout**: If Sentry enabled, slogmulti.Fanout writes to base handler + slogsentry (Error level only)
- **All code**: Uses stdlib `log/slog` directly (no third-party logging imports in business logic)

Sentry receives: stack traces (AttachStacktrace: true), all slog Error/above events.

## Frontend Architecture (React 19 SPA)

### Build Pipeline & Embedding

```
web/src/ (React + TypeScript)
  ├─ pnpm build (Vite 8)
  └─ web/dist/ (hashed assets: index.html, js/*, css/*)
      └─ make web
         └─ copy to internal/web/spa/public/
            └─ go build (//go:embed all:public)
               └─ static binary with SPA embedded
```

**Build Details**:
- **Vite 8**: Fast bundler with code splitting, lazy loading, HMR
- **Output**: `web/dist/` with hashed filenames for cache busting
- **Embedding**: `internal/web/spa/spa.go` uses `//go:embed all:public`
- **Asset Serving**: `/assets/*` served with 1-year immutable cache; `/` serves `index.html` (no-cache) for SPA shell

### Routing & Layouts (TanStack Router v1)

**File-based routing** in `web/src/routes/`:
```
routes/
├── __root.tsx           # Root layout + outlet
├── login.tsx            # Public login page
└── _authed.tsx          # Auth-guarded layout with topbar
    ├── index.tsx        # Dashboard (/)
    ├── _authed/
    │   ├── people/      # People CRUD (index, new, $personId, $personId.edit)
    │   ├── journal/     # Journal (index, new, $entryId, $entryId.edit)
    │   ├── gifts/       # Gifts (index, new, $giftId, $giftId.edit)
    │   ├── reminders/   # Reminders (index, new, $reminderId, $reminderId.edit)
    │   ├── dates/       # Important dates list
    │   ├── audit/       # Audit log view
    │   ├── me/          # User profile (index, setup)
    │   └── settings/    # Settings hub (index, labels, relationship-types, security)
```

**Layout Pattern**: `_authed.tsx` acts as auth guard with shared topbar; all routes under it require authentication.

### Components & Styling

**Component Library**: shadcn/ui primitives (Button, Card, Input, Select, Dialog, Sheet, Table, etc.) restyled for Linear/Stripe minimal:
- **Accent**: Indigo-600 (#4f46e5)
- **Surfaces**: Zinc palette (white, #fafafa muted, #e4e4e7 borders)
- **Borders**: Hairline (1px) zinc-200; no box shadows
- **Radius**: 0.375rem (compact aesthetic)
- **Typography**: Inter primary, JetBrains Mono for numerics; font-weight 600 headings

**Navigation**: `topbar.tsx` sticky header (h-14, border-b):
- Desktop (md+): Full nav inline with indigo underline active state
- Mobile (<md): Hamburger menu toggle; nav items in collapsible sidebar

**Charts**: Recharts v3.8.1 with custom indigo/zinc theme for dashboard visualizations

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

## HTTP Server (Echo v5)

### Global Middleware Stack (applied in order)

| Middleware | Purpose |
|-----------|---------|
| `middleware.Recover()` | Catches panics, returns 500 |
| `middleware.RequestID()` | Attaches unique request ID header to each request |
| `middleware.Gzip()` | Response compression (deflate/gzip) |
| `middleware.RequestLogger()` | Structured access logging via slog |
| Sentry middleware | Auto-captures request errors → Sentry |

### Routes (mounted in `internal/web/route.go`)

```
/health                → GET (liveness probe, no auth)
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

**Journal**, **Gifts**, **Reminders**, **Dates** — similar CRUD patterns

**Labels** — CRUD

**Relationships** — GET types, POST/DELETE type defs

**Audit** — GET audit log entries

All state-changing calls (POST/PUT/PATCH/DELETE) require `X-Requested-With: kith-spa` header when authenticated by cookie.

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

**FileService Interface**:
- `SaveAvatar(personID, file, header)` → saves file, returns relative path
- `DeleteAvatar(personID, path)` → removes file and cleans up empty directories
- `GetAvatarPath(personID)` → returns base directory for person's avatars

**LocalFileService Implementation**:
- **Base directory**: Configured via `AVATAR_STORAGE_PATH` (default: `data/avatars`)
- **Directory structure**: `data/avatars/{personID}/{randomStr}-{sanitized-name}.{ext}`
- **Atomic writes**: Temp file → sync → rename (prevents partial uploads)
- **Path traversal prevention**: Validates clean path stays within base directory

**Security Controls**:
- **MIME validation**: Dual-check (HTTP header + magic number via `http.DetectContentType`)
- **Allowed types**: `image/jpeg`, `image/png`, `image/gif`, `image/webp`
- **Size limit**: 5MB per file (enforced at handler + service layer)
- **Filename sanitization**: Alphanumeric + dash/underscore only; max 50 chars
- **Random prefix**: 8-byte hex prefix prevents filename collisions and guessing

**Integration with People Service**:
- Service stores avatar metadata in database: `avatar_path`, `avatar_mime_type`, `avatar_size`, `avatar_uploaded_at`
- On upload: saves file first, then updates DB in transaction; rolls back file on DB error
- On delete: clears DB metadata, then removes file (best-effort cleanup)
- On replace: saves new file, updates DB, then deletes old file (old file survives DB errors)

## Database Layer

### SQLite Configuration

**File**: `internal/db/sqlite.go`

Connection settings:
- **Driver**: `modernc.org/sqlite` (pure Go, no CGO)
- **WAL mode**: Write-Ahead Log for concurrent readers without blocking writer
- **Foreign keys**: Enabled (PRAGMA foreign_keys=ON)
- **Synchronous**: NORMAL (safe with WAL; balance of durability vs speed)
- **MaxOpenConns**: 1 (serializes all writes per SQLite single-writer model)

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
| `0011_audit_log.sql` | audit_log table for entity change tracking (entity_type, entity_id, entity_name, action, actor_id, created_at) |
| `0012_gift.sql` | gift table with direction (gave/received), debt_type (owed/owe), person association, and image storage metadata |
| `0013_person_self.sql` | is_self column on person table with unique index for self-profile feature |
| `0014_person_last_contact.sql` | last_contact_at nullable timestamp column on person table with partial index |
| `0015_relationship_type.sql` | relationship_type table (name, reverse_name, self-FK inverse_type_id) + person_relationship junction table (from_person_id, to_person_id, relationship_type_id, notes) with UNIQUE and CHECK constraints |

**Loading**: `internal/db/migrations.go` — loads SQL files in order, tracks applied versions in schema_migrations table.

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

### Audit Logging

**Architecture**:
- Service: `internal/audit/Service` — logs all entity mutations (CREATE, UPDATE, DELETE)
- Integration: Injected into people, journal, labels, reminders, dates, work_history services
- Best-effort: Logging failures never block primary operations; errors logged as warnings only
- Actor attribution: `audit.WithActor(ctx, userID)` and `ActorFromCtx(ctx)` for context-based user tracking
- Storage: `audit_log` table (entity_type, entity_id, entity_name, action, actor_id, created_at)

**Usage**:
```go
s.auditSvc.Log(ctx, audit.EntityType("person"), id, "Alice Smith", audit.ActionCreated)
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
```

## Deployment

### Single Binary
- Compiled with `CGO_ENABLED=0` → static binary, no libc or runtime dependencies
- Embedded React SPA (`//go:embed all:public` in `internal/web/spa/spa.go`)
  - Vite builds React 19 SPA to `web/dist/`
  - `make web` copies `web/dist/` → `internal/web/spa/public/`
  - `go embed` compiles all assets into binary (hashed filenames for cache busting)
- Embedded migrations (SQL files compiled into binary)
- **Build pipeline**: `make web` (pnpm build → copy) → `make build` (CGO_ENABLED=0 go build)`
- **Asset serving**: `/assets/*` served with 1-year cache headers; `/` serves index.html with no-cache + CSP headers

### Container (multi-stage Dockerfile)
- Stage 1: `node:22-alpine` — `pnpm install --frozen-lockfile && pnpm build`
- Stage 2: `golang:1.26.2-alpine` — copies SPA into embed path, runs `go build`
- Stage 3: `gcr.io/distroless/static-debian12` — minimal runtime, non-root UID 65532
- `go.uber.org/automaxprocs` auto-sets GOMAXPROCS to match container CPU quota
- Database: Mount volume at `/data` for persistent storage

### Safety Features
- Automatic migrations on startup (if DB_AUTO_MIGRATE=true)
- Backup via VACUUM INTO (safe while server running)
- Restore with 30-second server-activity heuristic (prevents data loss)
