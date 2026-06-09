# Stage 1: Build the React SPA
FROM node:24-alpine AS spa-builder

ENV CI=true

WORKDIR /app/web

# Install pnpm via corepack
RUN corepack enable && corepack prepare pnpm@latest --activate

COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY web/ ./
RUN pnpm build

# Stage 2: Build the Go binary
FROM golang:1.26.4-alpine AS go-builder

WORKDIR /app

# Cache Go module downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy all Go source
COPY . .

# Copy the built SPA into the embed path
RUN mkdir -p internal/web/spa/public
COPY --from=spa-builder /app/web/dist/ internal/web/spa/public/

# Build the static binary
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o kith-pms ./cmd

# Stage 3a: Distroless runtime image (default)
FROM gcr.io/distroless/static-debian12 AS distroless

COPY --from=go-builder /app/kith-pms /kith-pms

USER 65532:65532

EXPOSE 8000

ENTRYPOINT ["/kith-pms", "serve"]

# Stage 3b: Debian slim runtime image
FROM debian:12-slim AS debian-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=go-builder /app/kith-pms /kith-pms

RUN useradd --uid 65532 --no-create-home --shell /usr/sbin/nologin nonroot
USER 65532:65532

EXPOSE 8000

ENTRYPOINT ["/kith-pms", "serve"]
