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
│  │ Web Layer (Echo v5 + React SPA)                          │   │
│  │  ├─ /health        → Liveness probe (no auth)            │   │
│  │  ├─ /v1/*          → JSON REST API (Bearer or cookie)    │   │
│  │  ├─ /assets/*      → Embedded SPA assets (1yr cache)     │   │
│  │  └─ /*             → index.html catch-all (SPA shell)    │   │
│  │                                                          │   │
│  │  SPA routes (client-side, TanStack Router):              │   │
│  │  ├─ /              → Dashboard                           │   │
│  │  ├─ /people/*      → People CRUD                         │   │
│  │  ├─ /journal/*     → Journal                             │   │
│  │  ├─ /gifts/*       → Gifts                               │   │
│  │  ├─ /reminders/*   → Reminders                           │   │
│  │  ├─ /dates         → Important dates                     │   │
│  │  ├─ /audit         → Audit log                           │   │
│  │  └─ /settings/*    → Labels, rel types, security         │   │
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
/v1/*                  → JSON REST API (see api package for full route list)
/assets/*              → Embedded SPA hashed assets (1-year immutable cache)
/favicon.*             → Embedded favicon files
/*                     → index.html catch-all (200, no-cache) — SPA handles routing
```

#### JSON API routes (`/v1/*`)

```
POST   /v1/auth/login            → create session (rate limited 5/15min)
POST   /v1/auth/logout           → destroy current session
POST   /v1/auth/logout-all       → destroy all sessions for user
GET    /v1/auth/me               → current user or 401
POST   /v1/auth/password         → change password (requires current)
GET    /v1/me                    → self profile person
POST   /v1/me/setup              → set self person
GET    /v1/people                → list + search
POST   /v1/people                → create
GET    /v1/people/:id            → detail
PUT    /v1/people/:id            → update
DELETE /v1/people/:id            → delete
POST   /v1/people/:id/avatar     → upload avatar (multipart, max 5MB)
GET    /v1/people/:id/avatar     → retrieve avatar binary
DELETE /v1/people/:id/avatar     → delete avatar
...   (journal, gifts, reminders, dates, labels, relationship-types, audit)
```

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
4. Set session cookie (httpOnly, SameSite=Lax; Secure when BEHIND_TLS=true)
   - Name: kith_session
   - Value: signed HMAC token
   - Path: /
   - MaxAge: SESSION_LIFETIME
   ↓
5. On subsequent /v1/* requests:
   - SessionOrBearer middleware: accepts either cookie or Authorization: Bearer token
   - Cookie path: validates HMAC, looks up session, verifies expiry, injects user context
   - Bearer path: validates static TOKEN_AUTH value (machine clients only)
   - State-changing calls (POST/PUT/PATCH/DELETE) require X-Requested-With: kith-spa
     header when authenticated by cookie (CSRF protection via custom header)
   ↓
6. Password change (POST /v1/auth/password):
   - Verify current password
   - Hash new password with Argon2id
   - Invalidate all sessions
   - Rate limited: 5 attempts per 15 minutes
   ↓
7. Logout (POST /v1/auth/logout):
   - Clear session in database
   - Clear session cookie
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
- Compiled with `CGO_ENABLED=0`
- No runtime dependencies (everything bundled)
- Embedded React SPA (`//go:embed all:public` in `internal/web/spa/spa.go`)
- Embedded migrations (SQL files compiled into binary)
- Build pipeline: `make web` (pnpm build → copy to `internal/web/spa/public`) then `go build`

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
