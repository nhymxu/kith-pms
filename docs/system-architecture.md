# System Architecture

## High-Level Topology

```
┌─────────────────────────────────────────────────────────────────┐
│  Single Binary: kith-pms (CGO_ENABLED=0, no runtime deps)       │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ CLI Layer (urfave/cli v3)                                │   │
│  │  ├─ api              → starts HTTP server                │   │
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
│  │  ├─ /labels/*      → Labels CRUD                         │   │
│  │  └─ /journal/*     → Journal CRUD + FTS5 search          │   │
│  └──────────────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ Service Layer (auth, people, labels, journal)            │   │
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
| `api` | Start HTTP server on port from `HOST:PORT` (default :8000) |
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
| `HOST` | string | `0.0.0.0` | Server bind address |
| `PORT` | int | `8000` | Server bind port |
| `DB_PATH` | string | `data/kith.db` | SQLite database file path |
| `DB_AUTO_MIGRATE` | bool | `true` | Apply pending migrations on server startup |
| `SESSION_SECRET` | string | *(required)* | Cookie signing secret (min 32 bytes) |
| `SESSION_LIFETIME` | duration | `720h` (30 days) | Session cookie expiry duration |
| `BEHIND_TLS` | bool | `false` | Set `true` when behind TLS proxy (marks cookies Secure) |
| `DEBUG` | bool | `false` | `true` → text logs + debug level |
| `SENTRY_DSN` | string | *(empty)* | Sentry DSN; omit to disable error reporting |

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
/labels                → GET (list), POST (create)
/labels/:id            → GET (detail), PUT (update), DELETE
/journal               → GET (list + FTS5 search), POST (create)
/journal/:id           → GET (detail), PUT (update), DELETE
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

## Entity Relationships

```
users (1)
  ├─ (1:N) user_sessions
  └─ (1:N) people (created_by)

people (1)
  ├─ (1:N) contacts (phone, email)
  ├─ (1:N) locations (address)
  ├─ (N:M) labels (via label_assignments)
  └─ (N:M) activities (via activity_links)

labels (1)
  └─ (N:M) people (via label_assignments)

activities (1)  [Journal entries]
  ├─ (N:M) people (via activity_links)
  └─ (1:1 virtual) activities_fts [FTS5 index]
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
