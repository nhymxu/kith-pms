# Stage 1: build
# Installs codegen tools, runs asset generation, then compiles the binary.
# Assumes generated files (templ, sqlc) are NOT pre-committed.
FROM golang:1.26.2 AS builder

WORKDIR /app

# Install templ and sqlc code generators.
RUN go install github.com/a-h/templ/cmd/templ@v0.2.778 && \
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0

# Download Tailwind CSS standalone CLI (linux/amd64).
RUN curl -fsSLo /usr/local/bin/tailwindcss \
    https://github.com/tailwindlabs/tailwindcss/releases/download/v3.4.17/tailwindcss-linux-x64 \
    && chmod +x /usr/local/bin/tailwindcss

# Cache Go module downloads separately from source.
COPY go.mod go.sum ./
RUN go mod download

# Copy all source.
COPY . .

# Run code generation steps.
# sqlc: skip gracefully if no sqlc config present.
RUN sqlc generate -f internal/db/sqlc.yaml 2>/dev/null || true
# templ: generate all component files, then remove the stub.
RUN templ generate ./internal/web/... && \
    rm -f internal/web/templates/templates_stub.go
# Tailwind: compile and minify CSS.
RUN tailwindcss \
    -i internal/web/templates/styles.css \
    -o internal/web/static/app.css \
    --minify

# Build the static binary.
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o kith-pms ./cmd

# Stage 2: minimal runtime image.
FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/kith-pms /kith-pms

# Run as non-root (distroless nonroot UID).
USER 65532:65532

EXPOSE 8000

ENTRYPOINT ["/kith-pms", "api"]
