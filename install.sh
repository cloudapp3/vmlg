#!/usr/bin/env bash
set -euo pipefail

REPO="cloudapp3/vmlg"
BINARY="vmlg"
VERSION=""
INSTALL_DIR="${VMLG_INSTALL_DIR:-}"
SYSTEM_INSTALL=0
PRINT_DIR=0
SKIP_VERIFY=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      VERSION="$2"
      shift 2
      ;;
    --dir)
      INSTALL_DIR="$2"
      shift 2
      ;;
    --system)
      SYSTEM_INSTALL=1
      shift
      ;;
    --print-install-dir)
      PRINT_DIR=1
      shift
      ;;
    --skip-verify)
      SKIP_VERIFY=1
      shift
      ;;
    *)
      echo "Unknown option: $1" >&2
      exit 2
      ;;
  esac
done

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

if [[ -z "$INSTALL_DIR" ]]; then
  if [[ "$SYSTEM_INSTALL" -eq 1 || "$(id -u)" -eq 0 ]]; then
    INSTALL_DIR="/usr/local/bin"
  else
    INSTALL_DIR="$HOME/.local/bin"
  fi
fi

AUTH_ARGS=()
TOKEN="${GITHUB_TOKEN:-${GH_TOKEN:-}}"
if [[ -n "$TOKEN" ]]; then
  AUTH_ARGS=(-H "Authorization: Bearer ${TOKEN}")
fi

if [[ -z "$VERSION" ]]; then
  VERSION="$(curl -fsSL "${AUTH_ARGS[@]}" "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')"
fi
VER_STRIPPED="${VERSION#v}"
ARCHIVE="${BINARY}_${VER_STRIPPED}_${OS}_${ARCH}.tar.gz"
BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

curl -fsSL "${AUTH_ARGS[@]}" "${BASE_URL}/${ARCHIVE}" -o "${TMP_DIR}/${ARCHIVE}"

if [[ "$SKIP_VERIFY" -eq 0 ]]; then
  curl -fsSL "${AUTH_ARGS[@]}" "${BASE_URL}/checksums.txt" -o "${TMP_DIR}/checksums.txt"
  (
    cd "$TMP_DIR"
    EXPECTED_LINE="$(grep " ${ARCHIVE}\$" checksums.txt)"
    if command -v sha256sum >/dev/null 2>&1; then
      echo "$EXPECTED_LINE" | sha256sum -c - >/dev/null
    elif command -v shasum >/dev/null 2>&1; then
      echo "$EXPECTED_LINE" | shasum -a 256 -c - >/dev/null
    else
      echo "Neither sha256sum nor shasum found; pass --skip-verify to bypass." >&2
      exit 1
    fi
  )
fi

tar -xzf "${TMP_DIR}/${ARCHIVE}" -C "$TMP_DIR"

if [[ "$SYSTEM_INSTALL" -eq 1 && "$(id -u)" -ne 0 ]]; then
  sudo mkdir -p "$INSTALL_DIR"
  sudo install -m 0755 "${TMP_DIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
else
  mkdir -p "$INSTALL_DIR"
  install -m 0755 "${TMP_DIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
fi

if [[ "$PRINT_DIR" -eq 1 ]]; then
  echo "$INSTALL_DIR"
else
  echo "Installed ${BINARY} ${VERSION} to ${INSTALL_DIR}/${BINARY}"
fi
