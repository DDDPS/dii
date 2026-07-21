#!/bin/bash
set -e

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
if [ "$OS" != "linux" ]; then
    echo "Error: Unsupported OS '$OS'."
    exit 1
fi

ARCH="$(uname -m)"
case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Error: Unsupported architecture '$ARCH'."
        exit 1
        ;;
esac

DOWNLOAD_URL="https://github.com/DDDPS/dii/releases/latest/download/dii-${OS}-${ARCH}"
echo "Downloading dii for ${OS}-${ARCH}..."


curl -fsSL "$DOWNLOAD_URL" -o dii


chmod +x dii


INSTALL_DIR=""
if [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
elif [ -d "$HOME/.local/bin" ] && [ -w "$HOME/.local/bin" ]; then
    INSTALL_DIR="$HOME/.local/bin"
elif [ -d "$HOME/bin" ] && [ -w "$HOME/bin" ]; then
    INSTALL_DIR="$HOME/bin"
else
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
fi

mv dii "$INSTALL_DIR/dii"
echo "dii installed successfully to $INSTALL_DIR/dii!"
echo "Run 'dii --help' for usage instructions."

case ":$PATH:" in
    *":$INSTALL_DIR:"*) ;;
    *)
        echo ""
        echo "Warning: '$INSTALL_DIR' is not currently in your PATH."
        echo "To use 'dii' from anywhere, add it to your PATH by running:"
        echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc && source ~/.bashrc"
        ;;
esac
