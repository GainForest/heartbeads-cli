#!/usr/bin/env bash
set -euo pipefail

# hb (heartbeads-cli) installer
# Usage: curl -fsSL https://raw.githubusercontent.com/gainforest/heartbeads-cli/main/install.sh | bash

REPO="gainforest/heartbeads-cli"
BINARY="hb"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

# Colors (disabled if not a terminal)
if [ -t 1 ]; then
  BOLD='\033[1m'
  GREEN='\033[0;32m'
  YELLOW='\033[0;33m'
  RED='\033[0;31m'
  CYAN='\033[0;36m'
  RESET='\033[0m'
else
  BOLD='' GREEN='' YELLOW='' RED='' CYAN='' RESET=''
fi

info()  { printf "${CYAN}>${RESET} %s\n" "$*"; }
ok()    { printf "${GREEN}>${RESET} %s\n" "$*"; }
warn()  { printf "${YELLOW}!${RESET} %s\n" "$*"; }
fail()  { printf "${RED}x${RESET} %s\n" "$*"; exit 1; }

# --- Detect OS and architecture ---
detect_platform() {
  OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
  ARCH="$(uname -m)"

  case "$OS" in
    linux)  OS="linux" ;;
    darwin) OS="darwin" ;;
    *)      fail "Unsupported OS: $OS" ;;
  esac

  case "$ARCH" in
    x86_64|amd64)  ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *)             fail "Unsupported architecture: $ARCH" ;;
  esac
}

# --- Get latest release tag ---
# Returns 0 if a version was found, 1 otherwise (so caller can fall back).
get_latest_version() {
  if command -v curl &>/dev/null; then
    VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/' || true)"
  elif command -v wget &>/dev/null; then
    VERSION="$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/' || true)"
  else
    fail "curl or wget required"
  fi

  [ -n "$VERSION" ]
}

# --- Download and install binary from GitHub release ---
# Returns 0 on success, 1 on failure (so caller can fall back to source).
install_binary() {
  TARBALL="${BINARY}_${VERSION#v}_${OS}_${ARCH}.tar.gz"
  URL="https://github.com/${REPO}/releases/download/${VERSION}/${TARBALL}"

  TMPDIR="$(mktemp -d)"
  trap 'rm -rf "$TMPDIR"' EXIT

  info "Downloading hb ${VERSION} for ${OS}/${ARCH}..."

  if command -v curl &>/dev/null; then
    curl -fsSL "$URL" -o "${TMPDIR}/${TARBALL}" 2>/dev/null || return 1
  else
    wget -q "$URL" -O "${TMPDIR}/${TARBALL}" 2>/dev/null || return 1
  fi

  info "Extracting..."
  tar -xzf "${TMPDIR}/${TARBALL}" -C "$TMPDIR" || return 1

  mkdir -p "$INSTALL_DIR"
  mv "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  chmod +x "${INSTALL_DIR}/${BINARY}"
}

# --- Fallback: build from source ---
install_from_source() {
  if ! command -v go &>/dev/null; then
    fail "No prebuilt binary available and Go is not installed. Install Go from https://go.dev/dl/"
  fi

  info "No prebuilt binary found. Building from source..."
  go install "github.com/${REPO}/cmd/hb@latest" || fail "go install failed"
  INSTALL_DIR="$(go env GOPATH)/bin"
}

# --- Check bd dependency ---
check_bd() {
  if ! command -v bd &>/dev/null; then
    warn "bd (beads) not found in PATH"
    warn "Install it from https://github.com/gainforest/beads"
    warn "hb requires bd to function"
  fi
}

# --- Check PATH ---
check_path() {
  case ":$PATH:" in
    *":${INSTALL_DIR}:"*) ;;
    *)
      echo ""
      warn "${INSTALL_DIR} is not in your PATH. Add it:"
      echo ""
      echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
      echo ""
      echo "  Add that line to your ~/.bashrc, ~/.zshrc, or ~/.profile"
      ;;
  esac
}

# --- Main ---
main() {
  echo ""
  printf "${BOLD}  hb installer${RESET}\n"
  echo ""

  detect_platform

  # Try release binary first, fall back to building from source
  if get_latest_version && install_binary; then
    : # binary installed successfully
  else
    install_from_source
  fi

  check_bd

  # Verify installation
  HB_PATH="${INSTALL_DIR}/${BINARY}"
  if [ -x "$HB_PATH" ]; then
    HB_VERSION="$("$HB_PATH" --version 2>/dev/null || echo "unknown")"
    echo ""
    ok "hb installed successfully!"
    echo ""
    printf "  ${BOLD}%-12s${RESET} %s\n" "binary:" "$HB_PATH"
    printf "  ${BOLD}%-12s${RESET} %s\n" "version:" "$HB_VERSION"
    echo ""
    echo "  Get started:"
    echo ""
    echo "    hb account login --username <handle> --password <app-password>"
    echo "    hb init"
    echo "    hb ready"
    echo ""
  else
    fail "Installation failed: ${HB_PATH} not found"
  fi

  check_path
}

main
