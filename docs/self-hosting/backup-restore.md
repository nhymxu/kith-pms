# Backup & Restore

## Backup

The `backup` command uses SQLite's `VACUUM INTO` — safe to run while the server is running.

```bash
# Binary
./kith-pms backup --to /path/to/backup.db

# Docker
docker compose exec kith /kith-pms backup --to /data/backup-$(date +%Y%m%d).db
```

The backup file is a complete, standalone SQLite database.

## Restore

```bash
# Stop the server first, then:
./kith-pms restore --from /path/to/backup.db --force
```

The `--force` flag is required as a safety confirmation. The restore command refuses to proceed if the database was modified in the last 30 seconds (heuristic for a running server).

## Automated backups

Use a cron job or systemd timer to run backups on a schedule:

```bash
# Example cron — daily backup at 2am, keep 7 days
0 2 * * * /path/to/kith-pms backup --to /backups/kith-$(date +\%Y\%m\%d).db && find /backups -name 'kith-*.db' -mtime +7 -delete
```

## Verify a backup

```bash
sqlite3 /path/to/backup.db "PRAGMA integrity_check"
```
