#!/usr/bin/env bash
set -e

INSTALL_DIR="$HOME/.local/bin"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

mkdir -p "$INSTALL_DIR"
ln -sf "$SCRIPT_DIR/pacfind" "$INSTALL_DIR/pacfind"
chmod +x "$SCRIPT_DIR/pacfind"

echo "Pacfind installed successfully to $INSTALL_DIR/pacfind"
echo "You can now run 'pacfind <query>' from anywhere."
