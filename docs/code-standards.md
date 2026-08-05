# Code Standards

## Go Conventions

### Package Naming
- Lowercase, single word: `config`, `auth`, `audit`, `people`, `labels`, `journal`
- No underscores or mixed case in package names
- Package name matches directory name
- Domain packages under `internal/` (auth, audit, people, labels, journal, dates, reminders, files, gifts, work_history, relationships, settings, monica)
- Shared packages under `pkg/` (config, errors)

### File Naming
- Go standard: `snake_case.go` (e.g., `env.go`, `domain.go`, `service.go`, `repo.go`)
- Purpose patterns: `domain.go` (structs), `service.go` (business logic), `repo.go` (data access)
- Test files: `<name>_test.go` alongside the source file
- React components: `kebab-case.tsx` or `PascalCase.tsx` (e.g., `topbar.tsx`, `AppShell.tsx`, `dashboard-card.tsx`)

### Database & ORM Usage
- **ORM**: uptrace/bun (query builder + model mapping)
- **ORM Models**: Domain structs embed `bun.BaseModel` with `bun:"table:..."` tags; bun handles column mapping, time serialization, and struct scanning automatically
- **Query Builder**: Use `db.NewSelect()`, `db.NewInsert()`, `db.NewUpdate()`, `db.NewDelete()` throughout repos; raw SQL fragments allowed in `Where()`/`Join()`/`OrderExpr()` for FTS5, INTERSECT, and SQLite-specific functions
- **FTS5 Queries**: Journal full-text search uses raw SQL with bun: `db.NewRaw("SELECT ... WHERE rowid IN (SELECT rowid FROM activities_fts WHERE activities_fts MATCH ?)", term)`
- **Parameterized Queries**: Always use `?` placeholders — bun enforces this via its API; use `bun.List(slice)` for IN clauses
- **Transaction Pattern**: Write methods accept `bun.IDB` (satisfied by `*bun.DB` or `bun.Tx`) for unified transaction handling
- **Migration Files**: `0NNN_description.sql` in `internal/db/migrations/`; load programmatically in `internal/db/migrations.go`
- **Transactions**: Begin with `db.BeginTx(ctx, nil)`, always defer rollback, execute statements, commit when done

### Query Builder Pattern: Shared Builders Reused by List and Count
When a repo implements `List()` with pagination and filtering, extract the shared WHERE/JOIN logic into a private `buildXxxQuery(params *ListParams) *bun.SelectQuery` function reused by both:
- `List()` — add pagination (LIMIT, OFFSET) and ORDER BY to the base query
- `Count()` — uses the same base query but returns a COUNT(*) for total

**Example** (`internal/people/repo.go`):
```go
// Private builder; encapsulates shared WHERE/JOIN logic
func (r *repo) buildPeopleQuery(params *ListParams) *bun.SelectQuery {
    q := r.db.NewSelect().Model((*Person)(nil))
    if params.Query != "" {
        // Apply search filters
        q = q.Where("name LIKE ? OR nickname LIKE ?", "%"+params.Query+"%", "%"+params.Query+"%")
    }
    // Other filters (labels, dates, etc.)
    return q
}

// List uses builder + pagination
func (r *repo) List(ctx context.Context, params *ListParams) (*ListResult, error) {
    q := r.buildPeopleQuery(params)
    // Add pagination
    q = q.Limit(params.PageSize).Offset((params.Page - 1) * params.PageSize)
    q = q.Order("name ASC")
    var items []Person
    err := q.Scan(ctx, &items)
    return &ListResult{Items: items}, err
}

// Count uses same builder
func (r *repo) Count(ctx context.Context, params *ListParams) (int64, error) {
    q := r.buildPeopleQuery(params)
    return q.Count(ctx)
}
```

### Escaped-LIKE Search Convention (SQL Injection Prevention)
When exposing free-text search on user-controlled LIKE queries (people/gifts/etc.), escape the user input to neutralize `%`, `_`, and `\` metacharacters. Use `LIKE ? ESCAPE '\'`:

```go
// Private helper; escapes user input for LIKE queries
func escapeLikeQuery(userInput string) string {
    s := strings.Builder{}
    for _, r := range userInput {
        switch r {
        case '%', '_', '\\':
            s.WriteRune('\\')
        }
        s.WriteRune(r)
    }
    return s.String()
}

// Usage in list queries
func (r *repo) List(ctx context.Context, params *ListParams) (*ListResult, error) {
    escaped := escapeLikeQuery(params.Query)
    q := r.db.NewSelect().Model((*Person)(nil))
    if params.Query != "" {
        q = q.Where("name LIKE ? ESCAPE '\\'", "%"+escaped+"%")
    }
    // ...
}
```

**Why**: Raw LIKE patterns allow attackers to inject `%` (match any substring) or `_` (match single char) wildcards. The ESCAPE clause neutralizes these by treating the escaped char literally. Never trust user input in LIKE queries — always use this pattern.

### Struct Organization (Domain Models)
```go
// domain.go — data structures
type Person struct {
    ID          int64
    CreatedBy   int64
    Name        string
    DateOfBirth *time.Time
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// service.go — business logic
type Service struct {
    db bun.IDB
}

// repo.go — data access layer
type Repo struct {
    db *bun.DB
}
```

### Error Handling
- Wrap errors with context: `fmt.Errorf("operation failed: %w", err)`
- Never silently discard errors; log or propagate
- Panic only at startup for unrecoverable config failures
- All service methods return `error` as last return value

### Sentinel Error + Allowlist Validation Pattern
When a field has a fixed enum of valid values (theme, nav_layout, date_format, timezone, etc.), use a combination of:
1. **Sentinel errors** — one per invalid field (`ErrInvalidTheme`, `ErrInvalidNavLayout`, etc.)
2. **Allowlist map** — `validThemes map[string]bool`, `validNavLayout map[string]bool`, etc.
3. **Update validation in service** — call the map in `Service.Update()` before persisting
4. **Error mapping in handler** — switch on `errors.Is()` to return appropriate HTTP status (422 Unprocessable Entity)

**Example** (`internal/settings/service.go` + `internal/settings/domain.go`):
```go
// domain.go — sentinel errors and allowlist
var (
    ErrInvalidTheme = errors.New("settings: theme must be one of quiet-ink, warm-album, bold-press, nightdesk, softclay, ledger")
    ErrInvalidNavLayout = errors.New("settings: nav_layout must be one of top, side")
)

var validThemes = map[string]bool{
    "quiet-ink":  true,
    "warm-album": true,
    "bold-press": true,
    "nightdesk":  true,
    "softclay":   true,
    "ledger":     true,
}

var validNavLayout = map[string]bool{
    "top":  true,
    "side": true,
}

// service.go — validation before persist
func (s *Service) Update(ctx context.Context, updates *UserSettings) error {
    if !validThemes[updates.Theme] {
        return ErrInvalidTheme
    }
    if !validNavLayout[updates.NavLayout] {
        return ErrInvalidNavLayout
    }
    // Persist...
    return s.repo.Set(ctx, updates)
}

// handler.go — map sentinel errors to HTTP status
func (h *SettingsAPI) Update(c echo.Context) error {
    var req *settings.UserSettings
    c.Bind(&req)
    
    if err := h.Svc.Update(c.Request().Context(), req); err != nil {
        switch {
        case errors.Is(err, settings.ErrInvalidTheme):
            return apiErr(c, http.StatusUnprocessableEntity, err.Error())
        case errors.Is(err, settings.ErrInvalidNavLayout):
            return apiErr(c, http.StatusUnprocessableEntity, err.Error())
        default:
            return apiErr(c, http.StatusInternalServerError, "internal server error")
        }
    }
    return ok(c, req)
}
```

**Why**: Centralizes enum contract in one place (the allowlist); prevents typos in validation; makes the error message come from the canonical source, not duplicated across handlers.

### Logging
- Use `log/slog` (standard library) throughout — no third-party logging imports in business logic
- Structured key-value pairs: `slog.Info("msg", "key", value)`
- Log levels: `Debug` for dev detail, `Info` for lifecycle events, `Warn` for recoverable issues, `Error` for failures
- In production (`DEBUG=false`), slog outputs JSON; in debug mode, text format
- Use `slog.WithContext()` to pass request context through handlers
- **Sentry Integration**: Server-side only via `slog-sentry` fanout (optional, configured via `SENTRY.DSN`)
  - Never expose Sentry DSN in frontend bundles
  - Sentry receives Error level and above events with stack traces
  - Errors logged to slog automatically propagate to Sentry if configured

### Configuration Access
- All config consumed via `config.C` global (not `ENV`) — no direct `os.Getenv` calls outside `pkg/config`
- Add new config fields to `Config` struct in `pkg/config/env.go`
- Config loading via `config.Load()` function (not `LoadConfig`) using nhymxu/gommon/cfgloader
- Three-layer merge: hardcoded defaults → .env file → environment variables

### HTTP Handlers (Echo v5)
- Handlers live in `internal/api/handler/` package (one file per domain) with struct-based pattern
- **Handler Struct Pattern**: `type XxxAPI struct { Svc *xxx.Service }` with method receivers `(h *XxxAPI) Method(c echo.Context) error`
- Handler signatures: `func(h *XxxAPI) MethodName(c echo.Context) error` (Echo v5 method receiver pattern)
- Response helpers: Use `response.go` functions — `ok(c, data)`, `created(c, data)`, `apiErr(c, code, msg)` with {data, error} envelope
- Use `c.Bind()` for JSON/form binding to typed structs
- Use `c.QueryParam()`, `c.Param()` for individual values
- Response: JSON REST API only (SPA handles all UI rendering)
- CSRF middleware applied globally in `internal/api/server.go`; validates `X-Requested-With: kith-spa` header for state-changing calls

### Fan-Out / Aggregating Handler Exception
Most handlers depend on one service (e.g., `PeopleAPI` uses `PeopleSvc`). However, cross-domain endpoints that aggregate results from multiple services are legitimate:

**Example** (`SearchAPI` in `internal/api/handler/search.go`):
```go
// Aggregates queries to multiple services; NOT a code smell
type SearchAPI struct {
    PeopleSvc  *people.Service
    JournalSvc *journal.Service
    GiftsSvc   *gifts.Service
}

// Fans a single user query out to three services and assembles grouped results
func (h *SearchAPI) Search(c *echo.Context) error {
    peopleList, _ := h.PeopleSvc.List(ctx, people.ListParams{Query: q, PageSize: 5})
    journalList, _ := h.JournalSvc.List(ctx, journal.ListParams{Query: q, PageSize: 5})
    giftsList, _ := h.GiftsSvc.List(ctx, gifts.ListParams{Query: q, PageSize: 5})
    
    result := SearchResult{
        People:  toDTOs(peopleList.Items),
        Journal: toDTOs(journalList.Items),
        Gifts:   toDTOs(giftsList.Items),
    }
    return ok(c, result)
}
```

**Scope**: Use only for query (read-only) aggregation. Do NOT use for mutation handlers — keep writes focused per service.

### Middleware & Auth
- Register global middleware in `internal/api/server.go` (Recover, RequestID, Gzip, CSRF)
- Auth middleware checks session cookie, validates HMAC token, injects `*auth.User` into context
- Inject user into request: `c.Set("user", user)` — retrieve with `c.Get("user").(*auth.User)`
- CSRF validation automatic for POST/PUT/PATCH/DELETE when authenticated by cookie

### API Versioning
- All API routes prefixed with `/v1/` (e.g., `/v1/people`, `/v1/journal`)
- No deprecation policy yet; breaking changes require major version bump
- Health check at `/health` (no auth required)
- Readiness check at `/ready` (no auth required) — verifies DB connectivity and migrations applied
- Metrics at `/metrics` (no auth required) — Prometheus exposition format

### Transaction Patterns
```go
// Begin transaction
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return fmt.Errorf("begin tx: %w", err)
}
defer tx.Rollback() // Safe to call after Commit

// Execute via bun query builder on tx (satisfies bun.IDB)
_, err = tx.NewUpdate().Model(&person).WherePK().Column("name", "updated_at").Exec(ctx)
if err != nil {
    return fmt.Errorf("update failed: %w", err)
}

// Commit
if err := tx.Commit(); err != nil {
    return fmt.Errorf("commit failed: %w", err)
}
```

### Rate Limiting
- Login attempts: 5 per 15 minutes per IP (enforced in auth handler)
- Password changes: 5 per 15 minutes per user (enforced in auth handler)
- Implemented via in-memory rate limiter with sliding window

## React/TypeScript Frontend

### Routing
- **TanStack Router v1** file-based routing in `web/src/routes/`
- `_authed.tsx` layout pattern for auth guard
- Responsive mobile hamburger menu via `topbar.tsx`
- Route tree: 28 routes including `/` (dashboard), `/login`, `/people/*`, `/journal/*`, `/gifts/*`, `/reminders/*`, `/dates`, `/audit`, `/me/*`, `/settings/*`

### Components
- Functional components with hooks
- Use `#/` path alias for imports: `import { Button } from '#/components/ui/button'` (not `@/`)
- Shared primitives live in `web/src/components/ui`; use `@base-ui/react` for accessible primitive behavior when needed and preserve shadcn-style local component APIs
- Lucide React for icons only; no emojis

### Variant-Based Styling with CVA (class-variance-authority)
UI primitives that support multiple visual states (success, warning, danger, etc.) use `class-variance-authority` (CVA) for type-safe variant composition. This provides:
- Typescript-safe variant names (catches typos at build time)
- Composable, reusable style sets
- Self-documenting component prop signatures

**Example** (`web/src/components/ui/pill.tsx` — minimal badge component):
```typescript
import { cva, type VariantProps } from "class-variance-authority";

const pillVariants = cva(
    "inline-flex items-center gap-1 font-mono text-[10px] uppercase tracking-wide",
    {
        variants: {
            variant: {
                success: "text-success-fg",
                warning: "text-warning-fg",
                danger: "text-danger-fg",
                accent: "text-accent-text",
                plain: "text-sub",
            },
            strike: {
                true: "line-through",
                false: "",
            },
        },
        defaultVariants: {
            variant: "plain",
            strike: false,
        },
    }
);

export type PillVariant = NonNullable<VariantProps<typeof pillVariants>["variant"]>;

interface PillProps extends React.ComponentProps<"span">, VariantProps<typeof pillVariants> {}

function Pill({ className, variant, strike, ...props }: PillProps) {
    return <span className={cn(pillVariants({ variant, strike }), className)} {...props} />;
}
```

**Usage**: `<Pill variant="success">Done</Pill>`, `<Pill variant="warning" strike>Overdue</Pill>`

**Adoption**: Used across gifts-table (DebtBadge), audit-table (action-type badges), and reminders-table (StatusBadge) for consistent badge styling.

### Data Fetching
- **TanStack Query v5** with 5-minute stale time, 10-minute cache duration
- Define endpoints in `web/src/endpoints/*.ts` (e.g. `people.ts`, `journal.ts`, `gifts.ts`, `reminders.ts`)
- **Query hooks**: Use `useSuspenseQuery` (for routes/inner components within Suspense boundaries) or `useQuery` (for composite UI requiring `isLoading`/`isError` state like dashboard). Never mix patterns in the same component.
- **Suspense boundaries**: Wrap components using `useSuspenseQuery` with `<QueryBoundary>` (shared component in `web/src/components/query-boundary.tsx`) or `<Suspense>`. Shows fallback UI while queries load.
- **Error handling**: Route-level `errorComponent` (defined on `createFileRoute()`) catches suspended errors. Scoped `errorComponent` on detail routes shows domain-specific messages (e.g., "Gift not found."). Global `errorComponent` on `/_authed` layout handles auth-related errors.
- **Exception**: Dashboard intentionally uses `useQuery` to preserve composite KPI/chart state across queries; does not use Suspense.
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
- **Tailwind CSS 4.3.1** with Linear/Stripe minimal design tokens (Indigo-600 accent)
- Local UI components use indigo-600 accent, zinc surfaces, hairline borders, no shadows
- **Recharts 3.8+** for dashboard charts with custom Indigo/Zinc theme
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
- `make web` copies dist to `internal/api/spa/public/`
- Embedded into Go binary via `//go:embed all:public`

### First-Visit URL State Capture (`initial-search.ts`)

TanStack Router normalizes the URL bar with schema-defaulted search params before route components render, making it impossible to distinguish "user navigated to this URL with no params" from "params were filled in by validateSearch defaults" at route-level effects.

**Pattern**: `web/src/lib/initial-search.ts` exports `initialSearch = window.location.search` captured at module evaluation time (app boot, before router mount). Routes that need to detect a truly bare first visit import this instead of reading `window.location.search` from effects/components.

**Use case**: `/people` route applies settings-derived defaults (favorite-first, sort order, etc.) only on first visit, not on every re-render or back-navigation. It checks `new URLSearchParams(initialSearch).has("sort") || new URLSearchParams(initialSearch).has("favorite_only") || ...` to detect "user came with no explicit params" and fires a one-time redirect with merged defaults.

**Reusable for**: Any feature that needs "apply backend defaults only truly-once on bare first visit, then respect explicit URL params from that point forward."

### TypeScript Strict Mode
- Enable `strict: true` in `tsconfig.json`
- No `any` types without explicit `// @ts-ignore` comment with justification
- Use discriminated unions for type safety

### Biome Configuration
- Linter + formatter in `web/biome.json` (Biome 2.5.1+)
- Run `pnpm check` to verify lint/format
- Run `pnpm format` to auto-fix formatting issues
- Enforced on `make lint` via `pnpm --dir web check`

### Imports
- Group: stdlib → external → internal (separated by blank lines)
- Use the module path `github.com/nhymxu/kith-pms/...` for internal Go imports
- Standard Go imports: `"database/sql"`, `"time"`, `"fmt"`, `"log/slog"`
- External: `"github.com/labstack/echo/v5"`, `"golang.org/x/crypto/..."`

## File Storage Patterns

- **Avatar Storage**: `data/avatars/` with flat single-file scheme per person: `<personID>.<ext>` (JPEG, PNG, GIF, WebP; 32MB limit via `MAX_UPLOAD_SIZE_MB`; each person has exactly one avatar, new uploads replace old; on a format change `people.Service.UploadAvatar` deletes the previous file post-commit)
- **Gift Image Storage**: `data/gifts/` with filename pattern `<giftID>.<ext>` (JPEG, PNG, GIF, WebP; 32MB limit via `MAX_UPLOAD_SIZE_MB`; HEIC files converted to JPEG on client before upload)
- **Document Storage**: `data/documents/<personID>/` with original filename preserved (any file type; 50MB per file)
- All storage paths are configurable via environment variables (`AVATAR_STORAGE_PATH`, `GIFT_STORAGE_PATH`)
- MIME type detection at serve-time (no storage in DB); path traversal prevention in all methods
- **Image Retrieval Caching**: Avatar and gift images now use `Cache-Control: private, no-cache` with `ETag` for efficient 304 revalidation instead of 24-hour public caching

## Testing

### Go Backend Tests
- **Integration tests**: Use real SQLite database (e.g., `:memory:` or temp file)
- **Service tests**: `internal/{domain}/service_test.go` — test business logic with real repo
- **No mocks**: Prefer real dependencies over mocks for confidence in actual behavior
- **Test files**: 29 test files across auth, people, labels, journal, dates, files, reminders, relationships, gifts, work_history, settings, audit, metrics, monica
- **Total Go tests**: 200+ tests passing with race detector enabled

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
- **Embedding**: `make web` copies `web/dist/` to `internal/api/spa/public/` for Go embed

### Makefile Targets

| Target          | Command                                                         | Purpose                                      |
|-----------------|-----------------------------------------------------------------|----------------------------------------------|
| `web`           | `pnpm install && pnpm build && copy to internal/api/spa/public` | Build React SPA (Vite) and copy to embed dir |
| `swagger`       | `swag init -g cmd/doc.go -o internal/api/swagger`              | Generate OpenAPI 2.0 spec from swaggo annotations |
| `build`         | `make swagger && make web && CGO_ENABLED=0 go build ...`        | Full build (Swagger + SPA + static Go binary) |
| `dev`           | `make dev`                                                      | Run `go run ./cmd` with file watching        |
| `deps`          | `go mod download && go mod tidy`                                | Download and tidy Go dependencies            |
| `fmt`           | `gofmt -w .`                                                    | Auto-format Go files                         |
| `check-fmt`     | `gofmt -l . \| grep .`                                          | Verify Go formatting (fails if unformatted)  |
| `tidy`          | `gofmt -w . && go mod tidy`                                     | Format Go + tidy modules                     |
| `lint`          | `golangci-lint run ./...`                                       | Run Go linter                                |
| `tests`         | `go test -race ./...`                                           | Run all Go tests with race detector          |
| `test-coverage` | `go test -race -cover ./...`                                    | Go test coverage summary                     |
| `vuln-check`    | `govulncheck ./...`                                             | Scan Go for known vulnerabilities            |
| `gosec`         | `gosec ./...`                                                   | Go security analysis                         |

### Release & Distribution

**Goreleaser**: Multi-platform binary builds (Linux, macOS, Windows; amd64, arm64)
- Configuration: `.goreleaser.yml` in project root
- Build targets: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`, `windows/arm64`
- All builds use `CGO_ENABLED=0` for static binaries
- GitHub Actions workflow: Automated build + publish on git tag
- Artifacts: Pre-built binaries + SHA256 checksums on GitHub Releases

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

## Pagination & Page Size Configuration

### Page Size Precedence Pattern

Apply this pattern consistently across all paginated list views (people, journal, gifts, audit):

1. **Capture at Module Load**: `initial-search.ts` exports `initialSearch = window.location.search` at module evaluation time (before router mount) to detect bare first-visit
2. **Apply Defaults Once**: On first visit with no explicit `page_size` param, check user settings and redirect with merged defaults; on back-navigation, respect explicit URL params already present
3. **Precedence Chain** (in order):
   - URL `?page_size=N` (explicit user override)
   - localStorage `kith.page_size.<listKey>` (per-list remembered choice via `use-page-size-override.ts`)
   - Server `default_page_size` setting (user preference from `/v1/settings`)
   - Hardcoded fallback (20)
4. **UI Control**: `page-size-selector.tsx` dropdown with reset-to-default link; stores selection in localStorage per list
5. **Backend**: Server default in `DefaultPageSize` field (persisted via `user_setting` table); capped at 500

**Reusable for**: Any feature applying backend defaults only truly-once on bare first visit, then respecting explicit URL params thereafter.

### Image Crop Before Upload Pattern

For any image upload with user-facing crop/resize UI (avatars, gift images, etc.):

1. **Crop Dialog Component**: `web/src/components/image-crop-dialog.tsx` wraps `react-easy-crop` with fixed aspect ratio + pan/zoom controls
2. **Crop Utility**: `web/src/lib/crop-image.ts` exports `cropImageToBlob(imageSrc, cropArea, aspect)` and `blobToFile(blob, filename)` helpers; encodes to JPEG at configurable quality (`IMAGE_JPEG_QUALITY`) with max edge scaling (`IMAGE_MAX_EDGE_PX`)
3. **Upload Handler**: Before `POST /v1/people/:id/avatar`, open dialog; user confirms crop; convert blob to file; upload
4. **Constraints Definition**: Store per-domain constraints (aspect ratio, size/MIME limits) in focused modules:
   - Avatar: 1:1 aspect, 32MB limit via `MAX_UPLOAD_SIZE_MB`, JPEG/PNG/GIF/WebP only
   - Gift images: 4:3 aspect, 32MB limit via `MAX_UPLOAD_SIZE_MB`, JPEG/PNG/GIF/WebP only (via `gift-image-constraints.ts`)
5. **HEIC Support**: Client-side conversion via `heic-to/csp` (libheif WASM, ~2.9MB lazy-loaded chunk) converts HEIC→JPEG before crop dialog; applies to all 3 image inputs (avatar, gift create, gift edit)
6. **Server-Side**: Crop happens client-side only; server validates final file size, MIME, and stores at flat path (no random prefix for avatars)

**Reusable for**: Any image upload requiring user-controlled composition before persistence.

## File Upload Patterns

### Avatar Upload Flow
1. **Handler** (`internal/api/handler/people.go`):
   - Limit request body: `http.MaxBytesReader(w, r.Body, (config.C.MaxUploadSizeMB*1024*1024)+bufferBytes)` (configurable via `MAX_UPLOAD_SIZE_MB`, default 32MB + 1MB overhead)
   - Extract multipart file: `c.FormFile("avatar")`
   - Delegate to service: `h.Svc.UploadAvatar(ctx, personID, file, header)`

2. **Service** (`internal/people/service.go`):
   - Call FileService to save file (returns relative path)
   - Begin transaction; update person avatar metadata in DB
   - On success: commit, then delete the previous avatar file when the new path differs (format change); pruning is post-commit so a failed transaction cannot destroy the avatar the DB still references
   - On error: rollback transaction, delete new file

3. **FileService** (`internal/files/service.go`):
   - Validate file size against `MAX_UPLOAD_SIZE_MB` limit (configurable, default 32MB)
   - Read file header (512 bytes) for magic number check via `http.DetectContentType`
   - Validate MIME type (header + detected) against allowlist (JPEG, PNG, GIF, WebP only)
   - Determine file extension from MIME type
   - Write to temp file, sync, rename to final location (atomic write; prevents partial uploads)
   - Return relative path: `{personID}.{ext}`
   - Note: Flat scheme with no subdirectory, no random prefix, or filename sanitization — each person has exactly one avatar (new uploads replace old via deterministic naming)

### Gift Image Upload Flow
- Similar to avatar flow but stored in `GIFT_STORAGE_PATH` (default: `data/gifts`)
- File naming: Gift ID as filename (e.g., `123.jpg`)
- Endpoints: `POST /v1/gifts/:id/image`, `GET /v1/gifts/:id/image`, `DELETE /v1/gifts/:id/image`

### Document Storage Flow (Monica Import)
1. **Import Handler** (`cmd/monica_import.go`):
   - Decodes base64 dataURL from Monica export
   - Delegates to FileService: `filesSvc.SaveDocument(personID, data, originalName)`
   - Creates DOCUMENT-labelled journal entry linking document to person

2. **FileService** (`internal/files/service.go`):
   - `SaveDocument(personID, data, originalName)` — raw bytes, any MIME type, 50MB max
   - No MIME allowlist (unlike avatars); accepts any file type (PDF, Excel, Word, images, etc.)
   - Sanitizes filename; generates 8-byte random hex prefix for uniqueness
   - Stores in: `{DOCUMENT_STORAGE_PATH}/documents/{personID}/{randomStr}-{sanitized-name}.{ext}`
   - Returns relative path for DB storage

**Configuration**:
- Documents stored under `data/` directory (same base as avatars)
- Separate `documents/` subdirectory per person
- 50MB per file limit, hardcoded (avatars/gift images instead use `MAX_UPLOAD_SIZE_MB`, default 32MB)
- No MIME type validation — accepts all file types from trusted Monica imports

### Security Controls
- **MIME validation**: Dual-check (HTTP header + magic number) prevents spoofed uploads
- **Size limit**: Configurable via `MAX_UPLOAD_SIZE_MB` (default 32MB) enforced at handler + service layer
- **Path traversal prevention**: `filepath.Clean()` + prefix check ensures file stays in base directory
- **Filename sanitization**: Removes special chars; limits length to prevent filesystem issues
- **Atomic writes**: Temp file + sync + rename prevents partial/corrupted uploads
- **Metadata storage**: MIME type, size, upload timestamp stored in DB for audit trail
- **Stale file cleanup**: Avatar uploads with different formats delete old siblings to prevent orphaning

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
- **Password hashing**: Argon2id (golang.org/x/crypto/argon2) — Medium parameters (3 iterations, 65MB memory)
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
