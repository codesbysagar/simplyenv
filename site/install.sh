#!/usr/bin/env bash
set -e

# simplyenv universal installer for Linux and macOS
REPO="codesbysagar/simplyenv"
BASE_URL="https://simplyenv.codesbysagar.com/downloads"

echo "=================================================="
echo "  🚀 Installing simplyenv — Modern direnv Alternative"
echo "=================================================="

OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Linux)
    TARGET_OS="linux"
    ;;
  Darwin)
    TARGET_OS="darwin"
    ;;
  *)
    echo "❌ Error: Unsupported operating system: $OS"
    echo "For Windows, please download simplyenv-windows-amd64.zip directly from https://simplyenv.codesbysagar.com"
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64|amd64)
    TARGET_ARCH="amd64"
    ;;
  aarch64|arm64)
    TARGET_ARCH="arm64"
    ;;
  *)
    echo "❌ Error: Unsupported CPU architecture: $ARCH"
    exit 1
    ;;
esac

TARBALL="simplyenv-${TARGET_OS}-${TARGET_ARCH}.tar.gz"
DOWNLOAD_URL="${BASE_URL}/${TARBALL}"

echo "==> Downloading $TARBALL..."
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/$TARBALL"
elif command -v wget >/dev/null 2>&1; then
  wget -q "$DOWNLOAD_URL" -O "$TMP_DIR/$TARBALL"
else
  echo "❌ Error: curl or wget is required to download simplyenv."
  exit 1
fi

echo "==> Extracting binary..."
tar -xzf "$TMP_DIR/$TARBALL" -C "$TMP_DIR"

INSTALL_DIR="/usr/local/bin"
INSTALLED=0

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP_DIR/simplyenv" "$INSTALL_DIR/simplyenv"
  chmod +x "$INSTALL_DIR/simplyenv"
  INSTALLED=1
elif sudo -n true 2>/dev/null; then
  sudo mv "$TMP_DIR/simplyenv" "$INSTALL_DIR/simplyenv"
  sudo chmod +x "$INSTALL_DIR/simplyenv"
  INSTALLED=1
else
  INSTALL_DIR="$HOME/.local/bin"
  mkdir -p "$INSTALL_DIR"
  mv "$TMP_DIR/simplyenv" "$INSTALL_DIR/simplyenv"
  chmod +x "$INSTALL_DIR/simplyenv"
  INSTALLED=1
  
  if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
    echo "⚠️  Note: $HOME/.local/bin is not in your PATH. You may want to add:"
    echo "    export PATH=\"\$HOME/.local/bin:\$PATH\""
  fi
fi

echo ""
echo "✨ simplyenv installed successfully to $INSTALL_DIR/simplyenv!"
echo ""
echo "Next step: Run the 1-command installer to configure your shell hook:"
echo "    simplyenv install-hook"
echo ""
echo "Or launch the visual web dashboard anytime:"
echo "    simplyenv ui"
echo ""
