# Stage 1: Build the React SPA
FROM node:22-alpine AS spa-builder

WORKDIR /app/web

# Install pnpm via corepack (bundled with Node 22)
RUN corepack enable && corepack prepare pnpm@latest --activate

COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY web/ ./
RUN pnpm build

# Stage 2: Build the Go binary
FROM golang:1.26.2-alpine AS go-builder

WORKDIR /app

# Install sqlc for code generation
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0

# Cache Go module downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy all Go source
COPY . .

# Copy the built SPA into the embed path
RUN mkdir -p internal/web/spa/public
COPY --from=spa-builder /app/web/dist/ internal/web/spa/public/

# Generate sqlc query code (skip gracefully if no config)
RUN sqlc generate -f internal/db/sqlc.yaml 2>/dev/null || true

# Build the static binary
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o kith-pms ./cmd

# Stage 3: Minimal runtime image
FROM gcr.io/distroless/static-debian12

COPY --from=go-builder /app/kith-pms /kith-pms

# Run as non-root (distroless nonroot UID)
USER 65532:65532

EXPOSE 8000

ENTRYPOINT ["/kith-pms", "serve"]
