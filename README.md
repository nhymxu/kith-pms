# kith

**kith** (kith and kin) is a self-hosted personal relationship manager. Track people, tag them with labels, and keep a journal of interactions — all in a single binary backed by SQLite.

## Features

- **People** — store contacts with relationship type, date of birth, contact methods, addresses, and avatars
- **Self Profile** — designate one person as "Me" to filter journal entries and track personal participation
- **Labels** — colour-tagged categories; attach multiple labels to each person
- **Journal** — activity log with full-text search (SQLite FTS5); link entries to one or more people
- **Important Dates** — track birthdays, anniversaries, and milestones with "on this day" widget
- **Avatars** — upload profile pictures (JPEG, PNG, GIF, WebP; max 5MB) with automatic display and initials fallback
- **Relationships** — many-to-many person-to-person links with customizable, optionally-paired types (e.g. "Manager" / "Reports to")
- **Reminders** — scheduled reminders with due dates, person/date associations, and completion tracking
- **Dashboard** — at-a-glance counts and recent activity on the homepage
- **Auth** — single-user, password-protected; session cookies with CSRF protection
- **Backup / Restore** — safe online backup via SQLite `VACUUM INTO`; restore CLI with safety guard
- **Single binary** — no runtime dependencies; ships as a static binary (`CGO_ENABLED=0`)

## Quickstart

### Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.26+ | build |
| [templ](https://templ.guide) | v0.3.1001 | HTML component codegen |
| [Tailwind CSS CLI](https://tailwindcss.com/blog/standalone-cli) | v4.2+ | CSS build |
| make | any | convenience targets |

```bash
# Install templ
go install github.com/a-h/templ/cmd/templ@v0.3.1001

# Download Tailwind standalone CLI (macOS/Linux x64 — adjust for your platform)
curl -fsSLo tailwindcss \
  https://github.com/tailwindlabs/tailwindcss/releases/download/v4.2.4/tailwindcss-macos-x64
chmod +x tailwindcss && sudo mv tailwindcss /usr/local/bin/
```

### Build and run

```bash
git clone https://github.com/nhymxu/kith-pms.git
cd kith-pms

cp .env.example .env          # edit SESSION_SECRET (min 32 chars) and other vars

make assets                   # run templ generate + tailwind CSS build
make build                    # CGO_ENABLED=0 go build -o bin/kith-pms ./cmd

./bin/kith-pms migrate up     # create DB schema (data/kith.db by default)
./bin/kith-pms set-password   # set the login password interactively
./bin/kith-pms serve          # start the server on :8000
```

Open [http://localhost:8000](http://localhost:8000) and log in with the password you just set.

## Configuration

All configuration is via environment variables or a `.env` file in the working directory.

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `SESSION_SECRET` | — | **Yes** | Cookie signing secret, min 32 bytes |
| `DB_PATH` | `data/kith.db` | No | Path to the SQLite database file |
| `DB_AUTO_MIGRATE` | `true` | No | Run migrations automatically on startup |
| `AVATAR_STORAGE_PATH` | `data/avatars` | No | Directory for storing avatar files |
| `BEHIND_TLS` | `false` | No | Set `true` when serving behind TLS (marks cookies Secure) |
| `SESSION_LIFETIME` | `720h` (30 days) | No | How long a login session stays valid |
| `TOKEN_AUTH` | — | No | Bearer token for any future JSON API endpoints |
| `DEBUG` | `false` | No | `true` → text logs + debug level |
| `SENTRY_DSN` | — | No | Sentry DSN; omit to disable error reporting |

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

```bash
# Build and start
docker compose up -d

# Set password (first run)
docker compose exec kith /kith-pms set-password

# Backup
docker compose exec kith /kith-pms backup --to /data/backup.db
```

The `docker-compose.yml` mounts a named volume (`kith-data`) at `/data` for database persistence. Set `SESSION_SECRET` in your environment or a `.env` file before starting.

> **Note**: The Docker image runs as non-root (UID 65532). It uses `gcr.io/distroless/static-debian12` — no shell is available inside the container; use `docker logs` for debugging.

## Development

```bash
make deps          # download and tidy Go modules
make assets        # templ generate + tailwind CSS build
make build         # compile binary to bin/kith-pms
make fmt           # gofmt all Go files
make lint          # run golangci-lint
make tests         # run all tests with race detector
make test-coverage # generate coverage profile (HTML report)
make tidy          # fmt + go mod tidy
make vuln-check    # scan for known vulnerabilities
make gosec         # security static analysis
```

### Project layout

```
cmd/                CLI entrypoints (api, migrate, set-password, backup, restore)
internal/
  auth/             Session auth: users, sessions, middleware
  db/               SQLite open helper + embed migrations
  journal/          Activity journal domain, repo, service
  labels/           Labels domain, repo, service
  people/           People domain, repo, service
  dates/            Important dates & milestones
  reminders/        Reminders & notifications
  files/            File storage service (avatar uploads)
  web/              Echo HTTP server
    handlers/       HTTP handlers (auth, people, labels, journal, home, errors)
    templates/      templ components + CSS
    static/         Embedded static assets (htmx, compiled CSS)
pkg/
  config/           Config loading via koanf
docs/               Project documentation
plans/              Implementation plans
```

## Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.26, `CGO_ENABLED=0` |
| HTTP | Echo v5 |
| Database | SQLite (modernc.org/sqlite — pure Go) |
| Templates | [templ](https://templ.guide) |
| Interactivity | [htmx](https://htmx.org) |
| Styling | Tailwind CSS v4 |
| Auth | Argon2id password hash, signed cookie sessions |
| Search | SQLite FTS5 |

## Documentation

- [Project Overview](docs/project-overview-pdr.md)
- [Codebase Summary](docs/codebase-summary.md)
- [System Architecture](docs/system-architecture.md)
- [Code Standards](docs/code-standards.md)
- [Project Roadmap](docs/project-roadmap.md)
