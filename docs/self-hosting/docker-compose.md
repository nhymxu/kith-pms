# Docker Compose (Production)

## Prerequisites

- Docker Engine 24+ with Compose v2
- A domain name with DNS pointing to your server (for TLS)

## Setup

```bash
cd deploy/compose
cp .env.example .env
```

Edit `.env` and set at minimum:

```bash
# Generate with: openssl rand -hex 32
SESSION_SECRET=your-secret-here

# Set after first run: ./kith-pms set-password
APP_PASSWORD_HASH=your-bcrypt-hash
```

## Start

```bash
docker compose --env-file .env up -d
```

The app binds to `127.0.0.1:8000` by default. Put nginx or Caddy in front for TLS.

## Set password (first run)

```bash
docker compose exec kith /kith-pms set-password
```

Then restart: `docker compose restart kith`

## Backup

```bash
docker compose exec kith /kith-pms backup --to /data/backup-$(date +%Y%m%d).db
```

Copy the backup file out of the container:

```bash
docker compose cp kith:/data/backup-20260519.db ./backup-20260519.db
```

## Upgrade

```bash
docker compose pull
docker compose up -d
```

## Security notes

- `/metrics` is unauthenticated — do not expose port 8000 directly to the internet
- Set `BEHIND_TLS=true` when behind a TLS-terminating proxy to enable the Secure cookie flag
