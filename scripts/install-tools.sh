#!/usr/bin/env bash
# install-tools.sh — installs pinned dev tooling for kith-pms.
#
# Tools installed:
#   sqlc   v1.27.0   (via go install)
#   templ  v0.2.778  (via go install)
#   tailwindcss v3.4.17  standalone CLI → bin/tailwindcss
#
# Usage: bash scripts/install-tools.sh
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="${REPO_ROOT}/bin"

mkdir -p "${BIN_DIR}"

# ── sqlc ──────────────────────────────────────────────────────────────────────
echo "→ Installing sqlc v1.27.0 ..."
go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0
echo "  sqlc: $(sqlc version 2>/dev/null || echo 'installed')"

# ── templ ─────────────────────────────────────────────────────────────────────
echo "→ Installing templ v0.2.778 ..."
go install github.com/a-h/templ/cmd/templ@v0.2.778
echo "  templ: $(templ version 2>/dev/null || echo 'installed')"

# ── Tailwind CSS standalone CLI v3.4.17 ───────────────────────────────────────
TAILWIND_VERSION="3.4.17"
TAILWIND_BIN="${BIN_DIR}/tailwindcss"

# Detect OS and architecture to pick the correct release asset.
OS="$(uname -s)"
ARCH="$(uname -m)"

case "${OS}" in
  Linux)
    case "${ARCH}" in
      x86_64)  TAILWIND_ASSET="tailwindcss-linux-x64" ;;
      aarch64) TAILWIND_ASSET="tailwindcss-linux-arm64" ;;
      armv7l)  TAILWIND_ASSET="tailwindcss-linux-armv7" ;;
      *)       echo "Unsupported Linux arch: ${ARCH}"; exit 1 ;;
    esac
    ;;
  Darwin)
    case "${ARCH}" in
      x86_64)  TAILWIND_ASSET="tailwindcss-macos-x64" ;;
      arm64)   TAILWIND_ASSET="tailwindcss-macos-arm64" ;;
      *)       echo "Unsupported macOS arch: ${ARCH}"; exit 1 ;;
    esac
    ;;
  MINGW*|MSYS*|CYGWIN*)
    TAILWIND_ASSET="tailwindcss-windows-x64.exe"
    TAILWIND_BIN="${BIN_DIR}/tailwindcss.exe"
    ;;
  *)
    echo "Unsupported OS: ${OS}"
    exit 1
    ;;
esac

TAILWIND_URL="https://github.com/tailwindlabs/tailwindcss/releases/download/v${TAILWIND_VERSION}/${TAILWIND_ASSET}"

echo "→ Downloading tailwindcss v${TAILWIND_VERSION} (${TAILWIND_ASSET}) ..."
curl -fsSL --retry 3 -o "${TAILWIND_BIN}" "${TAILWIND_URL}"
chmod +x "${TAILWIND_BIN}"
echo "  tailwindcss: $("${TAILWIND_BIN}" --help 2>&1 | head -1 || echo 'installed')"

echo ""
echo "All tools installed successfully."
echo "  sqlc       → $(which sqlc)"
echo "  templ      → $(which templ)"
echo "  tailwindcss → ${TAILWIND_BIN}"
