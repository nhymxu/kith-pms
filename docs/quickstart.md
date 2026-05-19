# Quickstart

## Binary (recommended)

Download the latest release binary for your platform, then:

```bash
# Set a password
./kith-pms set-password

# Start the server (default port 8000)
./kith-pms serve
```

Open [http://localhost:8000](http://localhost:8000) and log in.

## Docker

```bash
cd deploy/compose
cp .env.example .env
# Edit .env — set SESSION_SECRET (openssl rand -hex 32) and APP_PASSWORD_HASH
docker compose --env-file .env up -d
```

## Build from source

```bash
git clone https://github.com/nhymxu/kith-pms
cd kith-pms
make build          # builds web SPA + Go binary → bin/kith-pms
./bin/kith-pms set-password
./bin/kith-pms serve
```

## Configuration

All config is via environment variables or a `.env` file:

| Variable            | Default        | Description                                         |
|---------------------|----------------|-----------------------------------------------------|
| `SESSION_SECRET`    | —              | Required. ≥32 random bytes (`openssl rand -hex 32`) |
| `APP_PASSWORD_HASH` | —              | Required. Set via `set-password` command            |
| `DB_PATH`           | `data/kith.db` | SQLite database path                                |
| `PORT`              | `8000`         | HTTP listen port                                    |
| `BEHIND_TLS`        | `false`        | Set `true` when behind a TLS proxy                  |
| `SESSION_LIFETIME`  | `720h`         | Session cookie lifetime                             |
