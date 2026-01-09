#!/bin/sh
set -e

REPO="kylemclaren/ralph"
BINARY_NAME="ralph"
INSTALL_DIR="/usr/local/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() { printf "${GREEN}[INFO]${NC} %s\n" "$1"; }
warn() { printf "${YELLOW}[WARN]${NC} %s\n" "$1"; }
error() { printf "${RED}[ERROR]${NC} %s\n" "$1"; exit 1; }

detect_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *) error "Unsupported operating system: $(uname -s)" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        armv7l) echo "arm" ;;
        *) error "Unsupported architecture: $(uname -m)" ;;
    esac
}

get_latest_version() {
    if command -v curl >/dev/null 2>&1; then
        curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    else
        return 1
    fi
}

download_file() {
    url="$1"
    output="$2"
    if command -v curl >/dev/null 2>&1; then
        curl -sL "$url" -o "$output"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$url" -O "$output"
    else
        error "Neither curl nor wget found. Please install one of them."
    fi
}

install_from_release() {
    OS=$(detect_os)
    ARCH=$(detect_arch)

    info "Detected OS: $OS, Arch: $ARCH"

    VERSION=$(get_latest_version)
    if [ -z "$VERSION" ]; then
        return 1
    fi

    info "Latest version: $VERSION"

    # Construct download URL (adjust pattern to match your release assets)
    if [ "$OS" = "windows" ]; then
        FILENAME="${BINARY_NAME}_${OS}_${ARCH}.zip"
    else
        FILENAME="${BINARY_NAME}_${OS}_${ARCH}.tar.gz"
    fi

    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

    info "Downloading from $DOWNLOAD_URL..."

    TMPDIR=$(mktemp -d)
    trap 'rm -rf "$TMPDIR"' EXIT

    if ! download_file "$DOWNLOAD_URL" "$TMPDIR/$FILENAME" 2>/dev/null; then
        return 1
    fi

    info "Extracting..."
    cd "$TMPDIR"
    if [ "$OS" = "windows" ]; then
        unzip -q "$FILENAME"
    else
        tar -xzf "$FILENAME"
    fi

    # Find the binary
    if [ -f "$BINARY_NAME" ]; then
        BINARY_PATH="$BINARY_NAME"
    elif [ -f "${BINARY_NAME}/${BINARY_NAME}" ]; then
        BINARY_PATH="${BINARY_NAME}/${BINARY_NAME}"
    else
        return 1
    fi

    install_binary "$TMPDIR/$BINARY_PATH"
}

install_from_go() {
    if ! command -v go >/dev/null 2>&1; then
        return 1
    fi

    info "Installing via go install..."
    go install "github.com/${REPO}/cmd/${BINARY_NAME}@latest"

    GOBIN=$(go env GOPATH)/bin
    if [ -f "$GOBIN/$BINARY_NAME" ]; then
        info "Installed to $GOBIN/$BINARY_NAME"

        # Check if GOBIN is in PATH
        case ":$PATH:" in
            *":$GOBIN:"*) ;;
            *)
                warn "Go bin directory is not in your PATH"
                warn "Add this to your shell profile:"
                echo ""
                echo "  export PATH=\"\$PATH:$GOBIN\""
                echo ""
                ;;
        esac
        return 0
    fi
    return 1
}

install_binary() {
    BINARY_PATH="$1"

    # Try to install to INSTALL_DIR
    if [ -w "$INSTALL_DIR" ]; then
        cp "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"
        chmod +x "$INSTALL_DIR/$BINARY_NAME"
        info "Installed to $INSTALL_DIR/$BINARY_NAME"
    elif command -v sudo >/dev/null 2>&1; then
        info "Installing to $INSTALL_DIR (requires sudo)..."
        sudo cp "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"
        sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
        info "Installed to $INSTALL_DIR/$BINARY_NAME"
    else
        # Fallback to user's local bin
        LOCAL_BIN="$HOME/.local/bin"
        mkdir -p "$LOCAL_BIN"
        cp "$BINARY_PATH" "$LOCAL_BIN/$BINARY_NAME"
        chmod +x "$LOCAL_BIN/$BINARY_NAME"
        info "Installed to $LOCAL_BIN/$BINARY_NAME"

        case ":$PATH:" in
            *":$LOCAL_BIN:"*) ;;
            *)
                warn "$LOCAL_BIN is not in your PATH"
                warn "Add this to your shell profile:"
                echo ""
                echo "  export PATH=\"\$PATH:$LOCAL_BIN\""
                echo ""
                ;;
        esac
    fi
}

main() {
    echo ""
    echo "  Ralph Installer"
    echo "  ==============="
    echo ""

    # Try release download first, then fall back to go install
    if install_from_release 2>/dev/null; then
        info "Installation complete!"
    elif install_from_go; then
        info "Installation complete!"
    else
        error "Installation failed. Please ensure you have Go installed or check https://github.com/${REPO}/releases"
    fi

    echo ""
    info "Run 'ralph help' to get started"
}

main "$@"
