# Migrating from Monica to kith

This guide covers exporting your data from Monica v4 and importing it into kith.

## Step 1 — Export from Monica

1. Log in to your Monica instance
2. Go to **Settings → Export data**
3. Choose **Export as JSON** and click **Export**
4. Wait for the export job to complete (you'll get an email or see a download link)
5. Download the `.json` file

The export is a single JSON file. It may be large (10 MB+) if you have many photos or documents — that's normal.

---

## Step 2 — Set up kith

If kith is not yet running:

```bash
git clone https://github.com/nhymxu/kith-pms.git
cd kith-pms
cp .env.example .env
# Edit .env — set SESSION_SECRET (min 32 chars)

CGO_ENABLED=0 go build -o bin/kith-pms ./cmd
./bin/kith-pms migrate up
./bin/kith-pms set-password
```

> If you also need the web interface, run `make build` instead (requires Node.js + pnpm for the SPA build).

If kith is already running, stop the server before importing to avoid lock contention:

```bash
# Stop the server, then run the import, then restart
```

---

## Step 3 — Dry run (preview)

Always do a dry run first to verify the file parses correctly:

```bash
./bin/kith-pms monica-import --from /path/to/monica-export.json --dry-run
```

If the export contains inactive reminders or account-level journal entries, the dry run uses the same default prompts as a real import. Pass explicit choices to make preview runs non-interactive:

```bash
./bin/kith-pms monica-import --from /path/to/monica-export.json --dry-run \
  --inactive-reminders=skip \
  --account-journal=skip
```

Example output:

```
Parsed 142 contacts from monica-export.json

Dry-run summary:
  Contacts (people):   142
  Contact info:        89
  Locations:           34
  Tags (labels):       12
  Journal entries:     310
  Reminders:           27
  Important dates:     98
  Gifts:               15
  Work history:        41
  Relationships:       63
```

No data is written. If the contact count looks wrong, check that you downloaded the full export (not a partial).

---

## Step 4 — Run the import

```bash
./bin/kith-pms monica-import --from /path/to/monica-export.json
```

By default, the import asks before including Monica data that cannot be linked cleanly:

| Option | Values | Default | Behavior |
|---|---|---|---|
| `--inactive-reminders` | `ask`, `skip`, `completed` | `ask` | `completed` imports inactive Monica reminders as completed kith reminders; `skip` omits them. |
| `--account-journal` | `ask`, `skip`, `unlinked` | `ask` | `unlinked` imports Monica v4 account-level journal entries without person links; `skip` omits them. |

For unattended imports, set both options explicitly:

```bash
./bin/kith-pms monica-import --from /path/to/monica-export.json \
  --inactive-reminders=completed \
  --account-journal=unlinked
```

The command logs each imported contact and finishes with a summary:

```
Import complete: 142 imported, 0 skipped/errors
```

Contacts skipped due to empty names or database errors are logged as warnings with the reason.

The import is safe to re-run on a fresh database. Running it twice on the same database will create duplicate records — restore from backup first if you need to re-import.

---

## Step 5 — Verify avatars

After import, avatars that were stored as photos in Monica (`avatar_source: photo`) are automatically saved to the `AVATAR_STORAGE_PATH` directory (default: `data/avatars`). The dry-run summary includes an `Avatars:` count showing how many photos will be imported. Check that the path exists and is writable before running the import if you expect avatars to be present.

Contacts whose avatars use gravatar, adorable, or external URLs will have no avatar in kith — upload one manually from their profile page.

## Step 6 — Verify all data

Start the server and check your data:

```bash
./bin/kith-pms serve
```

Open [http://localhost:8000](http://localhost:8000) and verify:

- People list has all your contacts
- Labels/tags are created and attached correctly
- Journal entries appear per person (notes + activities + calls)
- Important dates (birthdays, first-met) are present
- Reminders are listed (includes incomplete tasks from Monica)
- Gifts are listed per person
- Work history entries are present
- Relationships between contacts are linked
- Avatars appear for contacts that had locally-uploaded photos in Monica (the import summary prints `Avatars: N imported` if any were found)

---

## What gets migrated

| Monica field | kith field | Notes |
|---|---|---|
| first_name + last_name | Person name | |
| middle_name | Person name | Included between first and last |
| nickname | Nickname | |
| description | OtherNotes | Prepended before work info |
| job + company | Work history | Stored as a work history entry; start date defaults to 2000 since Monica doesn't export it |
| birthdate (with year) | Person DOB + Important date (birthday) | |
| birthdate (year unknown) | Important date only | Stored as `--MM-DD` |
| first_met_date | Important date (kind: met) | |
| addresses | Locations | name mapped to home/work/other |
| contact_fields | Contact info | `contact_fields.type` is a UUID and usually imports as `other` |
| tags | Labels | Created with colour `#6366f1` if new |
| notes | Journal entries | Note body becomes content; title truncated to 60 chars |
| activities | Journal entries | summary → title, description → content |
| calls | Journal entries | `Call on YYYY-MM-DD` title, content from call notes |
| reminders | Reminders | Active reminders import by default. Inactive reminders prompt by default; use `--inactive-reminders=completed` to import them as completed reminders or `--inactive-reminders=skip` to omit them. |
| tasks (incomplete only) | Reminders | Due date set to import day; completed tasks are skipped |
| gifts | Gifts | status mapped to given/received/planned; amount converted to cents |
| relationships | Person relationships | Relationship type names are created automatically |
| avatar (photo source) | Person avatar | Automatically imported when Monica `avatar_source` is `"photo"` and the photo data exists in the export. Photos are decoded from dataURL and saved to `AVATAR_STORAGE_PATH` (default: `data/avatars`). Gravatar, adorable, and external avatars are skipped. |

### What does NOT migrate

| Monica data | Reason |
|---|---|
| Gravatar / adorable / external avatars | Only locally-uploaded Monica photos (`avatar_source: photo`) are imported. Other avatar types have no file data in the export. |
| Documents | Same reason as photos — base64 embedded; no equivalent feature in kith. |
| Conversations / messages | No equivalent feature in kith. |
| Life events | No equivalent feature in kith. |
| Pets | No equivalent feature in kith. |
| Debts | No equivalent feature in kith (use gifts with debt_type if needed). |
| Completed tasks | Intentionally skipped — already done. |
| Journal entries (account-level) | Prompted by default because they have no person link; use `--account-journal=unlinked` to import them as unlinked journal entries or `--account-journal=skip` to omit them. |
| Audit logs | Internal Monica data, not relevant to kith. |
| Sync tokens / vCard data | Internal Monica data, not relevant to kith. |
| Inactive reminders | Prompted by default; use `--inactive-reminders=completed` to import them as completed reminders or `--inactive-reminders=skip` to omit them. |

---

## Troubleshooting

**"Parsed 0 contacts"** — The file may be a SQL export instead of JSON, or the download was incomplete. Re-export and choose JSON format.

**Many "skipped/errors"** — Run with verbose logging to see details:

```bash
DEBUG=true ./bin/kith-pms monica-import --from export.json
```

**Duplicate data after re-import** — The import does not deduplicate. Restore your database from backup before re-running:

```bash
./bin/kith-pms restore --from /path/to/backup.db --force
./bin/kith-pms monica-import --from export.json
```

**Work history start date shows 2000** — Monica's export does not include job start dates. Edit the work history entries manually in kith after import.

**Relationship types created with generic names** — Monica exports relationship type names as strings (e.g. "Friend", "Colleague"). These are created automatically in kith. Review and merge duplicates under **Settings → Relationship types** if needed.
