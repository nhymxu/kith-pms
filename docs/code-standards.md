# Code Standards

## Go Conventions

### Package Naming
- Lowercase, single word: `config`, `auth`, `audit`, `people`, `labels`, `journal`
- No underscores or mixed case in package names
- Package name matches directory name
- Domain packages under `internal/` (auth, audit, people, labels, journal, dates, reminders, files, db, web, api)
- Shared packages under `pkg/` (config)

### File Naming
- Go standard: `snake_case.go` (e.g., `env.go`, `domain.go`, `service.go`, `repo.go`)
- Purpose patterns: `domain.go` (structs), `service.go` (business logic), `repo.go` (data access)
- Test files: `<name>_test.go` alongside the source file
- React components: `kebab-case.tsx` or `PascalCase.tsx` (e.g., `topbar.tsx`, `AppShell.tsx`, `dashboard-card.tsx`)

### Database & SQL
- Use raw `database/sql` (no ORM)
- Parameterized queries only: `db.QueryRow("SELECT ... WHERE id = ?", id)` (no string concat)
- Migration files: `0NNN_description.sql` in `internal/db/migrations/` (currently 0001-0015)
- Load migrations programmatically in `internal/db/migrations.go`
- Transactions: Use `sql.Tx` for multi-statement operations; always defer rollback

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
- Handlers live in `internal/api/` (one file per domain)
- Handler signatures: `func(c echo.Context) error` (Echo v5 pattern)
- Return errors via `c.Error(err)` — let Echo's error handler format response
- Use `c.Bind()` for JSON/form binding to typed structs
- Use `c.QueryParam()`, `c.Param()` for individual values
- Response: JSON REST API only (SPA handles all UI rendering)
- CSRF middleware applied globally in `internal/web/server.go`; validates `X-Requested-With: kith-spa` header for state-changing calls

### Middleware & Auth
- Register global middleware in `internal/web/server.go` (Recover, RequestID, Gzip, CSRF)
- Auth middleware checks session cookie, validates HMAC token, injects `*auth.User` into context
- Inject user into request: `c.Set("user", user)` — retrieve with `c.Get("user").(*auth.User)`
- CSRF validation automatic for POST/PUT/PATCH/DELETE when authenticated by cookie

### API Versioning
- All API routes prefixed with `/v1/` (e.g., `/v1/people`, `/v1/journal`)
- No deprecation policy yet; breaking changes require major version bump
- Health check at `/health` (no auth required)

### Transaction Patterns
```go
// Begin transaction
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return fmt.Errorf("begin tx: %w", err)
}
defer tx.Rollback() // Safe to call after Commit

// Execute statements
_, err = tx.ExecContext(ctx, "UPDATE ...", args...)
if err != nil {
    return fmt.Errorf("update failed: %w", err)
}

// Commit
if err := tx.Commit(); err != nil {
    return fmt.Errorf("commit failed: %w", err)
}
```

## React/TypeScript Frontend

### Routing
- **TanStack Router v1** file-based routing in `web/src/routes/`
- `_authed.tsx` layout pattern for auth guard
- Responsive mobile hamburger menu via `topbar.tsx`
- Route tree: 28 routes including `/` (dashboard), `/login`, `/people/*`, `/journal/*`, `/gifts/*`, `/reminders/*`, `/dates`, `/audit`, `/me/*`, `/settings/*`

### Components
- Functional components with hooks
- Use `#/` path alias for imports: `import { Button } from '#/components/ui/button'` (not `@/`)
- shadcn/ui primitives restyled for Linear/Stripe minimal aesthetic
- Lucide React for icons only; no emojis

### Data Fetching
- **TanStack Query v5** with 5-minute stale time, 10-minute cache duration
- Define endpoints in `web/src/endpoints/*.ts` (e.g. `people.ts`, `journal.ts`, `gifts.ts`, `reminders.ts`)
- Use `useQuery` / `useMutation` hooks
- Query keys centralized in `web/src/query-keys.ts`

### Forms
- **TanStack Form v0** with Zod validation
- Define schemas in `web/src/schemas/` (hand-maintained per-resource, not generated)
- Schemas must align exactly with Go API types: case-sensitive field names, optional fields, enum values
- Example schema:
```typescript
export const PersonSchema = z.object({
  id: z.number(),
  name: z.string(),
  dateOfBirth: z.date().nullable(),
  labels: z.array(z.object({ id: z.number(), name: z.string() }))
});
```

### Styling
- **Tailwind CSS v4** with Linear/Stripe minimal design tokens
- **shadcn/ui** components restyled for indigo-600 accent, zinc surfaces, hairline borders, no shadows
- **Recharts v3.8.1** for dashboard charts
- Design tokens in `web/src/styles.css` (:root CSS variables)

### Authentication
- Consume via `lib/auth-context.tsx`
- Redirects to login if unauthenticated
- Session stored in `kith_session` HttpOnly cookie

### CSRF Protection
- All POST/PUT/PATCH/DELETE requests automatically include `X-Requested-With: kith-spa` header
- Handled in `lib/api-client.ts`

### Build
- **Vite 8** for bundling
- Output to `web/dist/`
- `make web` copies dist to `internal/web/spa/public/`
- Embedded into Go binary via `//go:embed all:public`

### TypeScript Strict Mode
- Enable `strict: true` in `tsconfig.json`
- No `any` types without explicit `// @ts-ignore` comment with justification
- Use discriminated unions for type safety

### Biome Configuration
- Linter + formatter in `web/biome.json`
- Run `pnpm check` to verify lint/format
- Run `pnpm format` to auto-fix formatting issues

### Imports
- Group: stdlib → external → internal (separated by blank lines)
- Use the module path `github.com/nhymxu/kith-pms/...` for internal Go imports
- Standard Go imports: `"database/sql"`, `"time"`, `"fmt"`, `"log/slog"`
- External: `"github.com/labstack/echo/v5"`, `"golang.org/x/crypto/..."`

## Testing

### Go Backend Tests
- **Integration tests**: Use real SQLite database (e.g., `:memory:` or temp file)
- **Service tests**: `internal/{domain}/service_test.go` — test business logic with real repo
- **No mocks**: Prefer real dependencies over mocks for confidence in actual behavior
- **Test files**: 15 test files across auth, people, labels, journal, dates, files, reminders, relationships, gifts
- **Total Go tests**: 159 tests passing with race detector enabled

### React Frontend Tests
- **Framework**: Vitest + @testing-library/react
- **Run tests**: `pnpm --dir web test`
- **Build checks**: `pnpm --dir web check` (Biome lint/format verification)

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
- **SPA Build**: `pnpm --dir web build` — Vite compiles React + TypeScript to `web/dist/`
- **CSS**: Tailwind CSS v4 (compiled via Vite plugin) using design tokens in `web/src/styles.css`
- **Embedding**: `make web` copies `web/dist/` to `internal/web/spa/public/` for Go embed

### Makefile Targets

| Target | Command | Purpose |
|--------|---------|---------|
| `web` | `pnpm install && pnpm build && copy to internal/web/spa/public` | Build React SPA (Vite) and copy to embed dir |
| `build` | `make web && CGO_ENABLED=0 go build -o bin/kith-pms ./cmd` | Full build (SPA + static Go binary) |
| `dev` | `make dev` | Run `go run ./cmd` with file watching |
| `deps` | `go mod download && go mod tidy` | Download and tidy Go dependencies |
| `fmt` | `gofmt -w .` | Auto-format Go files |
| `check-fmt` | `gofmt -l . \| grep .` | Verify Go formatting (fails if unformatted) |
| `tidy` | `gofmt -w . && go mod tidy` | Format Go + tidy modules |
| `lint` | `golangci-lint run ./...` | Run Go linter |
| `tests` | `go test -race ./...` | Run all Go tests with race detector |
| `test-coverage` | `go test -race -cover ./...` | Go test coverage summary |
| `vuln-check` | `govulncheck ./...` | Scan Go for known vulnerabilities |
| `gosec` | `gosec ./...` | Go security analysis |

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
- `chore: update dependencies, add Recharts v3.8.1`

No AI references in commit messages.

## File Upload Patterns

### Avatar Upload Flow
1. **Handler** (`internal/api/handlers_people.go`):
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

### Gift Image Upload Flow
- Similar to avatar flow but stored in `GIFT_STORAGE_PATH` (default: `data/gifts`)
- File naming: Gift ID as filename (e.g., `123.jpg`)
- Endpoints: `POST /v1/gifts/:id/image`, `GET /v1/gifts/:id/image`, `DELETE /v1/gifts/:id/image`

### Security Controls
- **MIME validation**: Dual-check (HTTP header + magic number) prevents spoofed uploads
- **Size limit**: 5MB enforced at handler + service layer
- **Path traversal prevention**: `filepath.Clean()` + prefix check ensures file stays in base directory
- **Filename sanitization**: Removes special chars; limits length to prevent filesystem issues
- **Atomic writes**: Temp file + sync + rename prevents partial/corrupted uploads
- **Metadata storage**: MIME type, size, upload timestamp stored in DB for audit trail

### Avatar Retrieval & Deletion
- **GET /v1/people/:id/avatar**: Validates path, sets Content-Type from DB, caches 24 hours
- **POST /v1/people/:id/avatar/delete**: Clears DB metadata, removes file (best-effort)

## Performance & Security Considerations

### Database
- **WAL mode**: Enables concurrent readers without blocking writer
- **MaxOpenConns=1**: Serializes writes per SQLite single-writer model
- **Prepared statements**: Use parameterized queries (?, not string concat)
- **FTS5**: Full-text search via virtual table with auto-update triggers

### Auth & Security
- **Password hashing**: bcrypt (golang.org/x/crypto/bcrypt)
- **Session tokens**: HMAC-SHA256 signed; server-stored with expiry
- **CSRF tokens**: Per-request tokens validated via middleware
- **Cookies**: Secure, httpOnly, SameSite=Lax
- **No secrets in logs**: Use structured logging with care for sensitive fields
- **File uploads**: MIME validation (header + magic number), size limits, path traversal prevention

### Deployment
- **Single binary**: All assets embedded; no external file dependencies
- **CGO_ENABLED=0**: Static binary; runs on any Linux/macOS/Windows (no libc dependency)
- **Backup safety**: VACUUM INTO is safe while server running
- **Migration safety**: Auto-applied on startup with version tracking
- **Avatar storage**: Configurable via AVATAR_STORAGE_PATH; ensure directory is writable and has sufficient disk space
- **Gift storage**: Configurable via GIFT_STORAGE_PATH; ensure directory is writable and has sufficient disk space
