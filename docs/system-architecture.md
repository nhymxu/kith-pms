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
│  │ Web Layer (Echo v5 + Templ + HTMX)                       │   │
│  │  ├─ /              → Dashboard (home)                    │   │
│  │  ├─ /auth/login    → Login form + session creation       │   │
│  │  ├─ /people/*      → People CRUD                         │   │
│  │  ├─ /settings/*    → Settings hub, labels, rel types     │   │
│  │  ├─ /journal/*     → Journal CRUD + FTS5 search          │   │
│  │  ├─ /dates         → Important dates & milestones        │   │
│  │  ├─ /reminders/*   → Reminders & notifications           │   │
│  │  ├─ /gifts/*       → Gifts CRUD + image upload/delete    │   │
│  │  └─ /audit         → Audit log with filter tabs          │   │
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

### Routes (mounted in `internal/web/server.go`)

```
/static/*              → Embedded assets (1-hour cache)
/                      → GET (dashboard)
/auth/login            → GET (form), POST (validate, create session)
/auth/logout           → POST (destroy session)
/people                → GET (list), POST (create)
/people/:id            → GET (detail), PUT (update), DELETE
/people/:id/edit       → GET (edit form)
/people/:id/date-row   → POST (add/update important date)
/people/:id/journal/quick → POST (quick-add journal entry, htmx fragment)
/people/:id/gifts/quick → POST (quick-add gift, htmx fragment)
/people/:id/avatar     → POST (upload), GET (retrieve)
/people/:id/avatar/delete → POST (delete)
/people/:id/last-contact → POST (update last contact timestamp to now)
/me                    → GET (self profile or setup redirect)
/me/setup              → GET (setup form), POST (set person as self)
/settings              → GET (settings hub with tiles for labels and rel types)
/settings/labels       → GET (list), POST (create)
/settings/labels/:id   → GET (detail), PUT (update), DELETE
/settings/relationship-types → GET (list, with counts), POST (create)
/settings/relationship-types/:id → GET (detail), POST (update), DELETE
/labels                → GET (302 redirect to /settings/labels)
/journal               → GET (list + FTS5 search), POST (create)
/journal/:id           → GET (detail), PUT (update), DELETE
/dates                 → GET (upcoming dates, ?days=N query param)
/reminders             → GET (list), POST (create)
/reminders/:id         → GET (detail), PUT (update), DELETE
/reminders/:id/complete → POST (mark as completed)
/audit                 → GET (paginated log, ?entity_type=X&page=N filters)
```

### Session & Auth Flow

```
1. User submits login form (POST /auth/login)
   ↓
2. Handler validates password (Argon2id verify vs users table)
   ↓
3. Create session token:
   - Generate cryptographic session ID
   - Store session in database (users.session_id, users.session_token, expires_at)
   - Sign session ID + user ID → HMAC-SHA256 token
   ↓
4. Set session cookie (secure, httpOnly, SameSite=Strict)
   - Value: signed HMAC token
   - Path: /
   - MaxAge: SESSION_LIFETIME
   ↓
5. On subsequent requests:
   - Middleware extracts session cookie
   - Validates HMAC signature
   - Looks up session in database
   - Verifies expiry time
   - Injects user context into request
   ↓
6. Logout (POST /auth/logout):
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
- Embedded static files (htmx.min.js, tailwind.css)
- Embedded migrations (SQL files compiled into binary)

### Container
- Dockerfile present; sets `CGO_ENABLED=0` at build stage
- `go.uber.org/automaxprocs` auto-sets GOMAXPROCS to match container CPU quota
- Database: Mount volume at `/app/data` for persistent storage

### Safety Features
- Automatic migrations on startup (if DB_AUTO_MIGRATE=true)
- Backup via VACUUM INTO (safe while server running)
- Restore with 30-second server-activity heuristic (prevents data loss)
