# Project Overview — kith-pms

## Vision

kith (kith and kin) is a self-hosted Personal Management System for individuals who want to track relationships, record life events, and retain personal memory — without the overhead of CRM tools designed for sales teams or enterprise workflows.

No collaboration features. No sharing. Just one person's data about the people and moments that matter to them.

## Target Users

Single individual user (self-hosted or personal deployment). No multi-tenancy in scope.

## Core Feature Areas (Implemented)

### Contacts & Relationships
- Store people: name, date of birth, relationship type, contact methods, addresses
- Self-profile designation: mark one person as "Me" for personal journal filtering
- Track interaction history via journal entries linked to people
- Tag relationships with color-coded labels
- Upload profile avatars (JPEG, PNG, GIF, WebP)
- Important dates tracking (birthdays, anniversaries, milestones)
- Reminders for follow-up and staying in touch
- Many-to-many person-to-person relationships with customizable, optionally-paired types
- Last contact timestamp tracking (manual & auto-update from journal entries)

### Life Journal / Log
- Date-stamped entries with title and content
- Link journal entries to multiple contacts
- Full-text search via SQLite FTS5
- Filter by date range and people
- Auto-update last contact for participants when self-profile is involved

### Memory & Notes
- Free-form journal entries tied to people
- Important dates with recurring support
- Searchable via FTS5 full-text index

### Timeline & Reminders
- "On this day" widget showing upcoming important dates
- Reminder system with due dates and completion tracking
- Link reminders to people or important dates

### Gifts & Money Tracking
- Track gifts given, received, and planned
- Money tracking per gift
- Gift photos/images with upload support
- Debt type tracking (owed/owe)

### Audit Log
- Automatic change tracking for all entities
- Timestamps and user attribution
- Full historical record of edits and deletions

## Tech Stack (Implemented)

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| Language | Go 1.26.2, CGO_ENABLED=0 | Compiled binary, low overhead, easy self-hosting |
| HTTP | Echo v5 | Lightweight HTTP framework, minimal magic |
| Database | SQLite (modernc.org/sqlite) | Pure Go, no CGO, single-file database |
| Frontend | React 19, TanStack Router v1 | CSR SPA with file-based routing; full client-side interactivity |
| Data Fetching | TanStack Query v5 | Cache-first data fetching, stale-while-revalidate, per-component refresh |
| Forms | TanStack Form v0 | Uncontrolled form state with Zod validation |
| Tables | TanStack Table v8 | Headless table library for data-heavy views |
| UI Components | shadcn/ui (Linear/Stripe minimal, restyled) | Headless component library with Tailwind theming |
| Styling | Tailwind CSS v4 | Utility-first CSS with design tokens |
| Build | Vite 6 | Fast bundler; code splitting, lazy loading, HMR |
| Linter/Formatter | Biome | Rust-based linter + formatter for JS/TS |
| Package Manager | pnpm | Fast, disk-efficient workspaces |
| CLI | urfave/cli v3 | Simple CLI scaffolding for subcommands |
| Config | koanf | Layered config: defaults → .env file → env vars |
| Logging | slog | Standard library structured logging (Go 1.21+) |
| Error Monitoring | slog-sentry | Fan-out errors to Sentry without replacing slog |
| Auth | Argon2id + HMAC sessions | Password hashing + signed HttpOnly cookie sessions |
| Search | SQLite FTS5 | Full-text search with auto-update triggers |

## Design System

Linear/Stripe minimal aesthetic: indigo-600 accent, zinc surfaces, Inter + JetBrains Mono typography, hairline borders, no shadows, responsive horizontal topbar. Built with React 19 CSR SPA and shadcn/ui components, styled via Tailwind CSS v4 design tokens.

## Non-Goals

- Multi-user / team features
- Mobile-native apps (web-first)
- Integration with external CRMs or calendars (at this stage)
- AI features (not in scope for initial phases)
