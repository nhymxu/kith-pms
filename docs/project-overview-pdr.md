# Project Overview — kith-pms

## Vision

kith (kith and kin) is a dead-simple Personal Management System for individuals who want to track relationships, record life events, and retain personal memory — without the overhead of CRM tools designed for sales teams or enterprise workflows.

No collaboration features. No sharing. Just one person's data about the people and moments that matter to them.

## Target Users

Single individual user (self-hosted or personal deployment). No multi-tenancy in scope.

## Core Feature Areas (Planned)

### Contacts & Relationships
- Store people: name, how you know them, notes
- Track interaction history (when you last talked, what about)
- Tag relationships by type (family, friend, colleague, etc.)
- Reminders to stay in touch

### Life Journal / Log
- Date-stamped entries — events, observations, milestones
- Link journal entries to contacts
- Search and filter by date range, tags, people

### Memory & Notes
- Free-form notes tied to people or topics
- Facts you want to remember (birthdays, preferences, context)
- Searchable, taggable

### Timeline
- Chronological view across all entries, contacts, and events
- "On this day" recall

## Tech Decisions Rationale

| Decision | Rationale |
|----------|-----------|
| Go | Compiled binary, low overhead, easy self-hosting |
| Echo v5 | Lightweight HTTP framework, minimal magic |
| urfave/cli v3 | Simple CLI scaffolding for subcommands (api, future tools) |
| koanf | Layered config: defaults → .env file → env vars, no reflection magic |
| slog | Standard library structured logging (Go 1.21+), no vendor lock-in |
| slog-sentry | Fan-out errors to Sentry without replacing slog |
| golangci-lint | Single tool for all lint rules |
| automaxprocs | Auto-tunes GOMAXPROCS in containerized environments |

## Design System

Frontend (planned) follows a PostHog-inspired design language: warm sage/olive palette, IBM Plex Sans typography, Tailwind CSS + Radix UI + shadcn/ui. Details in `DESIGN.md`.

## Non-Goals

- Multi-user / team features
- Mobile-native apps (web-first)
- Integration with external CRMs or calendars (at this stage)
- AI features (not in scope for initial phases)
