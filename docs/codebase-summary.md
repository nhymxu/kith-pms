# Codebase Summary

## Directory Structure

```
kith-pms/
├── cmd/                          # CLI entrypoints (compiled to single binary)
│   ├── main.go                   # Binary entry, dependency init, subcommand dispatch
│   ├── api.go                    # `api` subcommand — starts Echo HTTP server
│   ├── migrate.go                # `migrate` subcommand — runs schema migrations
│   ├── set_password.go           # `set-password` subcommand — interactive password setter
│   ├── backup.go                 # `backup` subcommand — database backup CLI
│   └── restore.go                # `restore` subcommand — database restore CLI
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
│   ├── people/                   # Contacts & relationships
│   │   ├── domain.go             # Person, Contact, Location structures
│   │   ├── service.go            # CRUD and query business logic
│   │   ├── repo.go               # Database queries
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
│   └── web/                      # HTTP handler layer
│       ├── server.go             # Echo setup, dependency injection, route mounting
│       ├── handlers/             # HTTP handlers for each domain
│       │   ├── auth.go           # Login, logout, password change
│       │   ├── home.go           # Dashboard / home page (includes OnThisDay widget)
│       │   ├── people.go         # CRUD handlers for People (dates integration)
│       │   ├── labels.go         # CRUD handlers for Labels
│       │   ├── journal.go        # CRUD handlers for Journal
│       │   ├── dates.go          # Handlers for Important Dates
│       │   └── errors.go         # Error page handlers
│       ├── templates/            # Templ HTML components (.templ files)
│       │   ├── layout.templ      # Base layout with navbar, footer
│       │   ├── login.templ       # Login form
│       │   ├── home.templ        # Dashboard (includes OnThisDay widget)
│       │   ├── people_list.templ, people_detail.templ, people_form.templ
│       │   ├── dates_list.templ  # Upcoming dates list
│       │   ├── labels_list.templ, labels_partials.templ
│       │   ├── journal_list.templ, journal_detail.templ, journal_form.templ
│       │   ├── journal_partials.templ
│       │   ├── error_404.templ, error_500.templ
│       │   ├── styles.css        # Tailwind CSS output
│       │   └── templates_stub.go # Templ code generation marker
│       ├── forms/                # Form validation & binding
│       │   └── forms.go          # Form struct definitions
│       └── static/               # Embedded static assets
│           └── htmx.min.js       # HTMX library for dynamic updates
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
│   └── 0007_important_date.sql   # Important dates table with virtual month_day column
├── scripts/
│   ├── lint.sh                   # Runs golangci-lint
│   ├── dependency-graph.sh       # Generates module dependency graph
│   └── find-cgo-pkg.sh           # Identifies CGO dependencies
├── web/                          # Frontend assets root (empty, for future use)
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
- **api.go**: `api` subcommand — starts Echo server after dependency init
- **migrate.go**: `migrate` subcommand — applies pending SQL migrations
- **set_password.go**: `set-password` subcommand — interactive password change
- **backup.go** & **restore.go**: Database backup/restore CLI with safety checks

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

### `internal/people` — Contacts management
- **domain.go**: Person (name, DOB, type), Contact (email, phone), Location (street, city, country)
- **service.go**: CRUD (CreatePerson, GetPerson, UpdatePerson, DeletePerson); query by label, search
- **repo.go**: Raw database/sql queries; JOIN queries for contacts & locations
- **service_test.go**: Integration tests for CRUD and complex queries

### `internal/labels` — Tag system
- **domain.go**: Label (name, color hex)
- **service.go**: CRUD; many-to-many label-person associations
- **repo.go**: Queries for label lists, person-label associations
- **service_test.go**: Integration tests for many-to-many logic

### `internal/journal` — Activity log with full-text search
- **domain.go**: Entry (title, content, date, author), EntryLink (link to People via FK)
- **service.go**: CRUD; full-text search via FTS5; link entries to multiple people
- **repo.go**: Queries including FTS5 search; maintains FTS5 trigger-based index
- **service_test.go**: Integration tests for FTS5 search

### `internal/dates` — Important dates & milestones
- **domain.go**: ImportantDate (kind, label, date_value, recurring), OnThisDayItem (date + person info)
- **service.go**: CRUD for dates; OnThisDay queries; Upcoming dates calculation
- **repo.go**: Queries for dates by person; OnThisDay matches; virtual month_day column queries
- **service_test.go**: Integration tests for date parsing, recurring logic, and queries

### `internal/web` — HTTP & template layer
- **server.go**: Creates Echo instance, mounts static file server, registers route groups, injects service dependencies into handlers
- **handlers/**: HTTP handler functions for each domain (auth, people, labels, journal, home)
- **templates/**: Templ HTML components (compiled to Go code); layouts, forms, detail pages, partials for HTMX swaps
- **forms/**: Form struct definitions for validation & binding
- **static/**: Embedded htmx library and generated Tailwind CSS

### `pkg/config` — Configuration
- **env.go**: LoadConfig() with three-layer merge (defaults → .env file → env vars); unmarshals to global ENV
- **const.go**: Application constants (timezones, timeouts, etc.)
- **default.go**: configDefaults map (lowest precedence layer)

## Key Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `labstack/echo/v5` | v5.1.0+ | HTTP framework |
| `urfave/cli/v3` | v3.8.0+ | CLI subcommands |
| `modernc.org/sqlite` | latest | Pure Go SQLite (no CGO) |
| `a-h/templ` | v0.2.778+ | HTML component codegen |
| `golang.org/x/crypto` | latest | Argon2id password hashing |
| `golang.org/x/text` | latest | Text encoding utilities |
| `knadh/koanf/v2` | v2.3.4+ | Layered config loading |
| `getsentry/sentry-go` | v0.46.1+ | Error monitoring (optional) |
| `samber/slog-multi` | v1.8.0+ | slog fan-out to multiple handlers |
| `samber/slog-sentry/v2` | v2.10.3+ | slog → Sentry integration |
| `go.uber.org/automaxprocs` | v1.6.0+ | Auto GOMAXPROCS in containers |

## Module & Build

- **Module**: `github.com/nhymxu/kith-pms` — Go 1.26.2+
- **Build**: CGO_ENABLED=0 for single static binary (no runtime dependencies)
- **Binary name**: `kith-pms` (compiled to `bin/kith-pms`)
- **Asset generation**: `make assets` → runs templ + tailwindcss (must run before `make build`)

## Test Coverage

33 tests across:
- `auth`: password hashing, session tokens, CSRF token generation
- `people`: CRUD, search, label associations
- `labels`: CRUD, many-to-many associations
- `journal`: CRUD, FTS5 full-text search

Run all: `make tests` (includes race detector)
