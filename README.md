# kith

**kith** (kith and kin) is a self-hosted personal relationship manager. Track people, tag them with labels, and keep a journal of interactions — all in a single binary backed by SQLite.

## Features

- **People** — store contacts with relationship type, date of birth, contact methods, addresses, and avatars
- **Self Profile** — designate one person as "Me" to filter journal entries and track personal participation
- **Labels** — colour-tagged categories; attach multiple labels to each person
- **Journal** — activity log with full-text search (SQLite FTS5); link entries to one or more people
- **Important Dates** — track birthdays, anniversaries, and milestones with "on this day" widget
- **Avatars** — upload profile pictures (JPEG, PNG, GIF, WebP; max 5MB) with automatic display and initials fallback
- **Relationships** — many-to-many person-to-person links with customizable, optionally-paired types; bulk connect-to-person action; network depth filter (direct/indirect contacts)
- **Relationship Network** — force-directed graph visualization of your entire contact network (global at `/network`; per-person ego view)
- **Reminders** — scheduled reminders with due dates, person/date associations, and completion tracking; 7 recurrence types including birthday reminders
- **Dashboard** — at-a-glance counts and recent activity on the homepage
- **Auth** — single-user, password-protected; session cookies with CSRF protection
- **Backup / Restore** — safe online backup via SQLite `VACUUM INTO`; restore CLI with safety guard
- **Single binary** — no runtime dependencies; ships as a static binary (`CGO_ENABLED=0`)

## Quickstart

### Prerequisites

| Tool     | Version  | Purpose                      |
|----------|----------|------------------------------|
| Go       | 1.26.4+  | build                        |
| Node.js  | 22 LTS   | React SPA build              |
| pnpm     | 11+      | JS package manager           |
| make     | any      | convenience targets          |

```bash
# Install pnpm (if not already installed)
npm install -g pnpm
# or via corepack (bundled with Node 22):
corepack enable && corepack prepare pnpm@latest --activate
```

### Build and run

```bash
git clone https://github.com/nhymxu/kith-pms.git
cd kith-pms

cp .env.example .env          # edit SESSION_SECRET (min 32 chars) and other vars

make web                      # pnpm install + pnpm build → copies SPA into internal/web/spa/public
make build                    # CGO_ENABLED=0 go build -o bin/kith-pms ./cmd
# (or just: make build — it runs make web automatically)

./bin/kith-pms migrate up     # create DB schema (data/kith.db by default)
./bin/kith-pms set-password   # set the login password interactively
./bin/kith-pms serve          # start the server on :8000
```

Open [http://localhost:8000](http://localhost:8000) and log in with the password you just set.

## Configuration

All configuration is via environment variables or a `.env` file in the working directory.

| Variable              | Default          | Required | Description                                               |
|-----------------------|------------------|----------|-----------------------------------------------------------|
| `SESSION_SECRET`      | —                | **Yes**  | Cookie signing secret, min 32 bytes                       |
| `DEBUG`               | `false`          | No       | `true` → text logs + debug level                          |
| `SENTRY.DSN`          | —                | No       | Sentry DSN; omit to disable error reporting               |
| `DB_PATH`             | `data/kith.db`   | No       | Path to the SQLite database file                          |
| `DB_AUTO_MIGRATE`     | `true`           | No       | Run migrations automatically on startup                   |
| `DB_MAX_OPEN_CONNS`   | `1`              | No       | SQLite connection pool size; default serializes writes    |
| `TOKEN_AUTH`          | —                | No       | Bearer token for any future JSON API endpoints            |
| `APP_PASSWORD_HASH`   | —                | No       | Pre-hashed Argon2id password for Docker/headless setup    |
| `AVATAR_STORAGE_PATH` | `data/avatars`   | No       | Directory for storing avatar files                        |
| `GIFT_STORAGE_PATH`   | `data/gifts`     | No       | Directory for storing gift image files                    |
| `BEHIND_TLS`          | `false`          | No       | Set `true` when serving behind TLS (marks cookies Secure) |
| `SESSION_LIFETIME`    | `720h` (30 days) | No       | How long a login session stays valid                      |

Environment variables take precedence over `.env` file values.

## Backup & Restore

### Backup

Creates a clean, compacted copy of the live database. Safe to run while the server is running.

```bash
./bin/kith-pms backup --to /path/to/backup.db
# Backed up data/kith.db → /path/to/backup.db  (1.2 MB → 1.1 MB)
```

> **Security**: the backup file contains all data including password hashes and session tokens.
> Store backups encrypted and restrict file permissions.

### Restore

Replaces the live database with a backup. **Stop the API server first.**

```bash
# Stop the server, then:
./bin/kith-pms restore --from /path/to/backup.db --force
# Restored /path/to/backup.db → data/kith.db  (1.1 MB)
```

The `--force` flag is required as a safety confirmation. The restore command also refuses to proceed if the database was modified in the last 30 seconds (heuristic for a running server).

## Docker

### Production

```bash
cd deploy/compose
cp .env.example .env
# Edit .env — set SESSION_SECRET and APP_PASSWORD_HASH at minimum
docker compose --env-file .env up -d

# Set password (first run)
docker compose exec kith /kith-pms set-password

# Backup
docker compose exec kith /kith-pms backup --to /data/backup.db
```

The production compose binds to `127.0.0.1` by default — put a TLS-terminating reverse proxy (nginx, Caddy) in front before exposing to the internet.

### Local development

```bash
# From repo root
docker compose -f compose.dev.yml up -d
```

> **Note**: The Docker image runs as non-root (UID 65532). It uses `gcr.io/distroless/static-debian12` — no shell is available inside the container; use `docker logs` for debugging.

## Development

```bash
make deps          # download and tidy Go modules
make web           # pnpm build + copy SPA into internal/web/spa/public
make build         # make web + CGO_ENABLED=0 go build (full build)
make assets        # alias for make web + sqlc codegen
make swagger       # generate OpenAPI docs from swaggo annotations
make fmt           # gofmt all Go files
make lint          # run golangci-lint
make tests         # run all tests with race detector
make test-coverage # generate coverage profile (HTML report)
make tidy          # fmt + go mod tidy
make clean         # remove web/dist and internal/web/spa/public
make vuln-check    # scan for known vulnerabilities
make dev           # Go live-reload (air) + Vite HMR
make gosec         # security static analysis
```

### Local dev with hot-reload (recommended)

```bash
make dev
```

This runs:
- **air** — watches Go files, auto-rebuilds and restarts the API on `:8000`
- **Vite** — frontend dev server on `:3000` with HMR, proxying `/v1/*` to `:8000`

### Local dev (SPA + API separately)

```bash
# Terminal 1 — Go API server on :8000
CGO_ENABLED=0 go run ./cmd serve

# Terminal 2 — Vite dev server on :3000 (proxies /v1 to :8000)
cd web && pnpm dev
```

### Project layout

```
cmd/                CLI entrypoints (serve, migrate, set-password, backup, restore)
internal/
  auth/             Session auth: users, sessions, middleware
  db/               SQLite open helper + embed migrations
  journal/          Activity journal domain, repo, service
  labels/           Labels domain, repo, service
  people/           People domain, repo, service
  dates/            Important dates & milestones
  reminders/        Reminders & notifications
  files/            File storage service (avatar/document uploads)
  api/              Echo HTTP server with route mounting
    handler/        HTTP handlers organized by domain
    spa/            Embedded React SPA (//go:embed public)
  relationships/    Relationship graph, queries, service
  settings/         User preferences & configuration
pkg/
  config/           Config loading via nhymxu/gommon/cfgloader
web/                React SPA source (pnpm workspace)
  src/              TanStack Router + Query + Form + shadcn/ui
  public/           Static assets (favicon, manifest)
docs/               Project documentation
```


## Documentation

- [Project Overview](docs/project-overview-pdr.md)
- [Codebase Summary](docs/codebase-summary.md)
- [System Architecture](docs/system-architecture.md)
- [Code Standards](docs/code-standards.md)

## Contributing & Development

### AI skills

manage AI skills via [skill.fish](https://github.com/knoxgraeme/skillfish) and store bundle on `skillfish.json`

`skill.fish` auto detect coding agent to install

```shell
mkdir -p .claude
npx skillfish install
```
