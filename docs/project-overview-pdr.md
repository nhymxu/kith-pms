# Project Overview — kith-pms

## Vision

kith (kith and kin) is a self-hosted Personal Management System for individuals who want to track relationships, record life events, and retain personal memory — without the overhead of CRM tools designed for sales teams or enterprise workflows.

No collaboration features. No sharing. Just one person's data about the people and moments that matter to them.

## Target Users

Single individual user (self-hosted or personal deployment). No multi-tenancy in scope.

## Core Feature Areas (Implemented)

### Contacts & Relationships
- Store people: name, date of birth, relationship type, contact methods, addresses
- Track interaction history via journal entries linked to people
- Tag relationships with color-coded labels
- Upload profile avatars (JPEG, PNG, GIF, WebP)
- Important dates tracking (birthdays, anniversaries, milestones)
- Reminders for follow-up and staying in touch

### Life Journal / Log
- Date-stamped entries with title and content
- Link journal entries to multiple contacts
- Full-text search via SQLite FTS5
- Filter by date range and people

### Memory & Notes
- Free-form journal entries tied to people
- Important dates with recurring support
- Searchable via FTS5 full-text index

### Timeline & Reminders
- "On this day" widget showing upcoming important dates
- Reminder system with due dates and completion tracking
- Link reminders to people or important dates

## Tech Stack (Implemented)

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| Language | Go 1.26, CGO_ENABLED=0 | Compiled binary, low overhead, easy self-hosting |
| HTTP | Echo v5 | Lightweight HTTP framework, minimal magic |
| Database | SQLite (modernc.org/sqlite) | Pure Go, no CGO, single-file database |
| Templates | templ | Type-safe HTML components compiled to Go |
| Interactivity | htmx | Dynamic UI without heavy JavaScript |
| Styling | Tailwind CSS v4 | Utility-first CSS with standalone CLI |
| CLI | urfave/cli v3 | Simple CLI scaffolding for subcommands |
| Config | koanf | Layered config: defaults → .env file → env vars |
| Logging | slog | Standard library structured logging (Go 1.21+) |
| Error Monitoring | slog-sentry | Fan-out errors to Sentry without replacing slog |
| Auth | Argon2id + HMAC sessions | Password hashing + signed cookie sessions |
| Search | SQLite FTS5 | Full-text search with auto-update triggers |

## Design System

Frontend follows a clean, minimal design: Tailwind CSS v4 for styling, htmx for interactivity, templ for type-safe HTML components.

## Non-Goals

- Multi-user / team features
- Mobile-native apps (web-first)
- Integration with external CRMs or calendars (at this stage)
- AI features (not in scope for initial phases)
