# Code Standards

## Go Conventions

### Package Naming
- Lowercase, single word: `config`, `auth`, `audit`, `people`, `labels`, `journal`
- No underscores or mixed case in package names
- Package name matches directory name
- Domain packages under `internal/` (auth, audit, people, labels, journal, dates, reminders, files, db, web)
- Shared packages under `pkg/` (config)

### File Naming
- Go standard: `snake_case.go` (e.g., `env.go`, `domain.go`, `service.go`, `repo.go`)
- Purpose patterns: `domain.go` (structs), `service.go` (business logic), `repo.go` (data access)
- Test files: `<name>_test.go` alongside the source file
- Templ files: `<domain>.templ` (e.g., `people_list.templ`, `people_form.templ`)

### Database & SQL
- Use raw `database/sql` (no ORM)
- Parameterized queries only: `db.QueryRow("SELECT ... WHERE id = ?", id)` (no string concat)
- Migration files: `0NNN_description.sql` in `internal/db/migrations/` (currently 0001-0011)
- Load migrations programmatically in `internal/db/migrations.go`

### Struct Organization (Domain Models)
```go
// domain.go — data structures
type Person struct {
    ID              int64
    CreatedBy       int64
    Name            string
    DateOfBirth     *time.Time
    RelationshipType string
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

// service.go — business logic
type Service struct {
    db *sql.DB
}

// repo.go — data access layer
type Repo struct {
    db *sql.DB
}
```

### Error Handling
- Wrap errors with context: `fmt.Errorf("operation failed: %w", err)`
- Never silently discard errors; log or propagate
- Panic only at startup for unrecoverable config failures
- All service methods return `error` as last return value

### Logging
- Use `log/slog` (standard library) throughout — no third-party logging imports in business logic
- Structured key-value pairs: `slog.Info("msg", "key", value)`
- Log levels: `Debug` for dev detail, `Info` for lifecycle events, `Warn` for recoverable issues, `Error` for failures
- In production (`DEBUG=false`), slog outputs JSON; in debug mode, text format
- Use `slog.WithContext()` to pass request context through handlers

### Configuration Access
- All config consumed via `config.ENV` global — no direct `os.Getenv` calls outside `pkg/config`
- Add new config fields to `EnvConfigMap` in `pkg/config/env.go` and defaults in `pkg/config/default.go`

### HTTP Handlers (Echo v5)
- Handlers live in `internal/web/handlers/` (one file per domain)
- Handler signatures: `func(c echo.Context) error` (Echo v5 pattern)
- Return errors via `c.Error(err)` — let Echo's error handler format response
- Use `c.Bind()` for form/query binding to typed structs
- Use `c.QueryParam()`, `c.Param()` for individual values
- Templ components in `internal/web/templates/` — receive context where needed
- CSRF middleware applied globally in `internal/web/server.go`; extract token via `c.Get("csrf_token")`

### Middleware & Auth
- Register global middleware in `internal/web/server.go` (Recover, RequestID, Gzip, CSRF)
- Auth middleware checks session cookie, validates HMAC token, injects `*auth.User` into context
- Inject user into request: `c.Set("user", user)` — retrieve with `c.Get("user").(*auth.User)`
- CSRF validation automatic for POST/PUT/DELETE via middleware

### Templ Components
- Extend base layout: `templ ComponentName(user *auth.User, data DataStruct) { ... }`
- Pass `c.Response().Header()` to set HTTP headers from templates
- Use htmx attributes: `hx-get`, `hx-post`, `hx-swap`, `hx-target` for dynamic updates
- Partials: `templ PartialName(data) { ... }` — called from handlers for HTMX swaps

### Imports
- Group: stdlib → external → internal (separated by blank lines)
- Use the module path `github.com/nhymxu/kith-pms/...` for internal imports
- Standard imports: `"database/sql"`, `"time"`, `"fmt"`, `"log/slog"`
- External: `"github.com/labstack/echo/v5"`, `"golang.org/x/crypto/..."`

## Testing

### Test Organization
- **Integration tests**: Use real SQLite database (e.g., `:memory:` or temp file)
- **Service tests**: `internal/{domain}/service_test.go` — test business logic with real repo
- **No mocks**: Prefer real dependencies over mocks for confidence in actual behavior
- **Test files**: 9 test files across auth, people, labels, journal, dates, files, reminders

### Test Structure
```go
func TestServiceCRUD(t *testing.T) {
    // Setup: create temp DB, seed schema
    db := setupTestDB(t)
    defer db.Close()
    
    svc := NewService(db)
    
    // Test: call service method
    id, err := svc.Create(ctx, &CreateInput{...})
    if err != nil {
        t.Fatalf("Create failed: %v", err)
    }
    
    // Verify: query database directly
    var created *Model
    err = db.QueryRow("SELECT ... WHERE id = ?", id).Scan(&created.ID, ...)
    if err != nil {
        t.Fatalf("Verify failed: %v", err)
    }
}
```

### Test Naming
- `TestServiceMethod` — service business logic
- `TestRepoQuery` — repository queries
- `TestPasswordHashVerify` — crypto functions

### Run Tests
```bash
make tests              # all tests with race detector
make test-coverage     # generate coverage report
```

## Build & Deployment

### CGO_ENABLED=0 Requirement
- Always build with `CGO_ENABLED=0` for static binary (no runtime deps)
- Used: `modernc.org/sqlite` (pure Go SQLite)
- Verified: `./scripts/find-cgo-pkg.sh` identifies any CGO dependencies

### Asset Generation
- **Templ**: `templ generate` — compiles `.templ` files to Go code
- **Tailwind**: `tailwindcss -i input.css -o styles.css` — builds CSS from template classes
- Run before build: `make assets`

### Makefile Targets

| Target | Command | Purpose |
|--------|---------|---------|
| `assets` | `templ generate && tailwindcss build` | Generate templates & CSS |
| `build` | `CGO_ENABLED=0 go build -o bin/kith-pms ./cmd` | Compile static binary |
| `deps` | `go mod download && go mod tidy` | Download and tidy dependencies |
| `fmt` | `gofmt -w .` | Auto-format all Go files |
| `check-fmt` | `gofmt -l . \| grep .` | Verify formatting (fails if unformatted) |
| `tidy` | `gofmt -w . && go mod tidy` | Format + tidy modules |
| `lint` | `golangci-lint run ./...` | Run linter |
| `tests` | `go test -race ./...` | Run tests with race detector |
| `test-coverage` | `go test -race -cover ./...` | Coverage summary |
| `vuln-check` | `govulncheck ./...` | Scan for known vulnerabilities |
| `gosec` | `gosec ./...` | Security analysis |

## Pre-commit Checklist

1. `make fmt` — format code
2. `make lint` — no lint errors
3. `make tests` — all tests pass
4. No `.env` or secrets committed
5. Database migrations properly numbered and tested

## Commit Messages

Use conventional commits:
- `feat:` new feature or capability
- `fix:` bug fix
- `refactor:` code restructure, no behavior change
- `test:` test additions/changes
- `chore:` tooling, deps, config, non-functional changes
- `docs:` documentation only

Examples:
- `feat: add FTS5 full-text search for journal entries`
- `fix: validate HMAC token before session lookup`
- `refactor: extract person repository from service`
- `test: add password hashing test vectors`
- `chore: update dependencies, add templ v0.3.1001`

No AI references in commit messages.

## File Upload Patterns

### Avatar Upload Flow
1. **Handler** (`internal/web/handlers/people.go`):
   - Limit request body: `http.MaxBytesReader(w, r.Body, 6*1024*1024)` (5MB file + 1MB overhead)
   - Extract multipart file: `c.FormFile("avatar")`
   - Delegate to service: `h.Svc.UploadAvatar(ctx, personID, file, header)`

2. **Service** (`internal/people/service.go`):
   - Call FileService to save file (returns relative path)
   - Begin transaction; update person avatar metadata in DB
   - On success: commit, then delete old avatar file (best-effort)
   - On error: rollback transaction, delete new file

3. **FileService** (`internal/files/service.go`):
   - Validate file size against limit (5MB)
   - Read file header (512 bytes) for magic number check via `http.DetectContentType`
   - Validate MIME type (header + detected) against allowlist
   - Sanitize filename: alphanumeric + dash/underscore; max 50 chars
   - Generate random 8-byte hex prefix to prevent collisions
   - Write to temp file, sync, rename (atomic write)
   - Return relative path: `{personID}/{randomStr}-{sanitized-name}.{ext}`

### Security Controls
- **MIME validation**: Dual-check (HTTP header + magic number) prevents spoofed uploads
- **Size limit**: 5MB enforced at handler + service layer
- **Path traversal prevention**: `filepath.Clean()` + prefix check ensures file stays in base directory
- **Filename sanitization**: Removes special chars; limits length to prevent filesystem issues
- **Atomic writes**: Temp file + sync + rename prevents partial/corrupted uploads
- **Metadata storage**: MIME type, size, upload timestamp stored in DB for audit trail

### Avatar Retrieval & Deletion
- **GET /people/:id/avatar**: Validates path, sets Content-Type from DB, caches 24 hours
- **POST /people/:id/avatar/delete**: Clears DB metadata, removes file (best-effort)

## Performance & Security Considerations

### Database
- **WAL mode**: Enables concurrent readers without blocking writer
- **MaxOpenConns=1**: Serializes writes per SQLite single-writer model
- **Prepared statements**: Use parameterized queries (?, not string concat)
- **FTS5**: Full-text search via virtual table with auto-update triggers

### Auth & Security
- **Password hashing**: Argon2id (golang.org/x/crypto/argon2)
- **Session tokens**: HMAC-SHA256 signed; server-stored with expiry
- **CSRF tokens**: Per-request tokens validated via middleware
- **Cookies**: Secure, httpOnly, SameSite=Strict
- **No secrets in logs**: Use structured logging with care for sensitive fields
- **File uploads**: MIME validation (header + magic number), size limits, path traversal prevention

### Deployment
- **Single binary**: All assets embedded; no external file dependencies
- **CGO_ENABLED=0**: Static binary; runs on any Linux/macOS/Windows (no libc dependency)
- **Backup safety**: VACUUM INTO is safe while server running
- **Migration safety**: Auto-applied on startup with version tracking
- **Avatar storage**: Configurable via AVATAR_STORAGE_PATH; ensure directory is writable and has sufficient disk space
