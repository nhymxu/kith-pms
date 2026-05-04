#!/usr/bin/env bash
# install-tools.sh — installs pinned dev tooling for kith-pms.
#
# Tools installed:
#   sqlc   v1.27.0   (via go install)
#   templ  v0.3.1001 (via go install)
#   tailwindcss v4.2.4  standalone CLI → bin/tailwindcss
#
# Usage: bash scripts/install-tools.sh
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="${REPO_ROOT}/bin"

mkdir -p "${BIN_DIR}"

if command -v brew >/dev/null 2>&1; then
  echo "Homebrew detected. You can install these tools with brew if you prefer:"
  echo "  brew install sqlc templ tailwindcss"
  echo "Continuing with pinned project-local installs..."
  echo ""
fi

# ── sqlc ──────────────────────────────────────────────────────────────────────
echo "→ Installing sqlc v1.27.0 ..."
CGO_ENABLED=0 go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0
echo "  sqlc: $(sqlc version 2>/dev/null || echo 'installed')"

# ── templ ─────────────────────────────────────────────────────────────────────
echo "→ Installing templ v0.3.1001 ..."
go install github.com/a-h/templ/cmd/templ@v0.3.1001
echo "  templ: $(templ version 2>/dev/null || echo 'installed')"

# ── Tailwind CSS standalone CLI v4.2.4 ────────────────────────────────────────
TAILWIND_VERSION="4.2.4"
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
