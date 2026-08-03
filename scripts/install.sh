#!/usr/bin/env bash
# Script: install.sh
# Purpose: Downloads and installs the latest zswapctl binary according to system architecture.
#
# Key Components:
#   - Detects architecture (amd64 vs arm)
#   - Downloads latest binary from GitHub releases
#   - Installs binary to /usr/bin/zswapctl
#   - Prompts user to apply default configuration via zswapctl --install
#
# Dependencies:
#   - curl or wget
#   - sudo or root privileges
#
# Error Types:
#   - Unsupported architecture or failed download

set -e

REPO="Dlcuy22/zswapctl"
ARCH=$(uname -m)

case "$ARCH" in
    x86_64)
        BINARY="zswapctl-linux-amd64"
        ;;
    arm*|aarch64)
        BINARY="zswapctl-linux-arm"
        ;;
    *)
        echo "Error: Unsupported architecture: $ARCH" >&2
        exit 1
        ;;
esac

DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${BINARY}"
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

echo "Downloading ${BINARY} from latest release..."
if command -v curl >/dev/null 2>&1; then
    curl -sSL "$DOWNLOAD_URL" -o "$TMP_DIR/zswapctl"
elif command -v wget >/dev/null 2>&1; then
    wget -qO "$TMP_DIR/zswapctl" "$DOWNLOAD_URL"
else
    echo "Error: Neither curl nor wget is installed." >&2
    exit 1
fi

chmod +x "$TMP_DIR/zswapctl"

echo "Installing zswapctl to /usr/bin/zswapctl..."
if [ "$(id -u)" -eq 0 ]; then
    install -Dm755 "$TMP_DIR/zswapctl" /usr/bin/zswapctl
else
    sudo install -Dm755 "$TMP_DIR/zswapctl" /usr/bin/zswapctl
fi

echo "zswapctl binary successfully installed."

if [ -e /dev/tty ]; then
    read -p "Do you want to apply default configuration? (y/n): " response </dev/tty
else
    response="n"
fi

case "$response" in
    [yY][eE][sS]|[yY])
        if [ "$(id -u)" -eq 0 ]; then
            zswapctl --install
        else
            sudo zswapctl --install
        fi
        ;;
    *)
        echo "Skipping default configuration installation."
        ;;
esac
