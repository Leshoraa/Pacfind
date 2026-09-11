#!/usr/bin/env bash
set -e

INSTALL_DIR="$HOME/.local/bin"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Building pacfind (Go binary)..."
cd "$SCRIPT_DIR"
go build -ldflags="-s -w" -o pacfind .

mkdir -p "$INSTALL_DIR"
ln -sf "$SCRIPT_DIR/pacfind" "$INSTALL_DIR/pacfind"

echo "Pacfind installed successfully to $INSTALL_DIR/pacfind"
echo "Run 'pacfind -v' or 'pacfind <query>' from anywhere."
