# Migrating from Monica to kith

This guide covers exporting your data from Monica and importing it into kith. Both Monica v2/v3 and v4 export formats are supported.

## Step 1 — Export from Monica

### Monica v4 (self-hosted)

1. Log in to your Monica instance
2. Go to **Settings → Export data**
3. Choose **Export as JSON** and click **Export**
4. Wait for the export job to complete (you'll get an email or see a download link)
5. Download the `.json` file

### Monica v2/v3

Same path: **Settings → Export data → Export as JSON**.

The export is a single JSON file. It may be large (10 MB+) if you have many photos or documents — that's normal.

---

## Step 2 — Set up kith

If kith is not yet running:

```bash
git clone https://github.com/nhymxu/kith-pms.git
cd kith-pms
cp .env.example .env
# Edit .env — set SESSION_SECRET (min 32 chars)

make build
./bin/kith-pms migrate up
./bin/kith-pms set-password
```

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

The command logs each imported contact and finishes with a summary:

```
Import complete: 142 imported, 0 skipped/errors
```

Contacts skipped due to empty names or database errors are logged as warnings with the reason.

The import is safe to re-run on a fresh database. Running it twice on the same database will create duplicate records — restore from backup first if you need to re-import.

---

## Step 5 — Verify

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

---

## What gets migrated

### Monica v2/v3 and v4

| Monica field | kith field | Notes |
|---|---|---|
| first_name + last_name | Person name | |
| middle_name (v4) | Person name | Included between first and last |
| nickname | Nickname | |
| description (v4) | OtherNotes | Prepended before work info |
| job + company | OtherNotes | `Work: job at company` |
| job + company (v4) | Work history | Stored as a work history entry; start date defaults to 2000 since Monica doesn't export it |
| birthdate (with year) | Person DOB + Important date (birthday) | |
| birthdate (year unknown) | Important date only | Stored as `--MM-DD` |
| first_met_date | Important date (kind: met) | |
| addresses | Locations | name mapped to home/work/other |
| contactInformation / contact_fields | Contact info | email, phone, social (twitter/facebook/linkedin/github/instagram), other |
| tags | Labels | Created with colour `#6366f1` if new |
| notes | Journal entries | Note body becomes content; title truncated to 60 chars |
| activities | Journal entries | summary → title, description → content |
| calls (v4) | Journal entries | `Call on YYYY-MM-DD` title, content from call notes |
| reminders | Reminders | |
| tasks (v4, incomplete only) | Reminders | Due date set to import day; completed tasks are skipped |
| gifts (v4) | Gifts | status mapped to given/received/planned; amount converted to cents |
| relationships (v4) | Person relationships | Relationship type names are created automatically |

### What does NOT migrate

| Monica data | Reason |
|---|---|
| Photos / avatars | Monica embeds them as base64 in the export; kith stores files on disk. Upload avatars manually after import. |
| Documents | Same reason as photos. |
| Conversations / messages | No equivalent feature in kith. |
| Life events | No equivalent feature in kith. |
| Pets | No equivalent feature in kith. |
| Debts | No equivalent feature in kith (use gifts with debt_type if needed). |
| Completed tasks | Intentionally skipped — already done. |
| Journal entries (account-level, v4) | Monica's account-level journal has no per-person link; not imported. |
| Audit logs | Internal Monica data, not relevant to kith. |
| Sync tokens / vCard data | Internal Monica data, not relevant to kith. |
| Inactive reminders (v4) | Skipped — marked inactive in Monica. |

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

**Relationship types created with generic names** — Monica v4 exports relationship type names as strings (e.g. "Friend", "Colleague"). These are created automatically in kith. Review and merge duplicates under **Settings → Relationship types** if needed.
