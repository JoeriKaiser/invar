#!/bin/sh
set -e

REPO="JoeriKaiser/invar"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY="invar"

main() {
    os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    arch="$(uname -m)"

    case "$arch" in
        x86_64|amd64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *) echo "Unsupported architecture: $arch" >&2; exit 1 ;;
    esac

    case "$os" in
        linux)  os="linux" ;;
        darwin) os="darwin" ;;
        *)      echo "Unsupported OS: $os" >&2; exit 1 ;;
    esac

    latest="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d '"' -f 4)"
    if [ -z "$latest" ]; then
        echo "Failed to fetch latest release." >&2
        exit 1
    fi

    filename="${BINARY}_${latest#v}_${os}_${arch}.tar.gz"
    url="https://github.com/${REPO}/releases/download/${latest}/${filename}"

    tmpdir="$(mktemp -d)"
    trap 'rm -rf "$tmpdir"' EXIT

    echo "Downloading ${BINARY} ${latest} for ${os}/${arch}..."
    curl -fsSL "$url" -o "${tmpdir}/${filename}"
    tar -xzf "${tmpdir}/${filename}" -C "$tmpdir"

    if [ -w "$INSTALL_DIR" ]; then
        mv "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    else
        echo "Installing to ${INSTALL_DIR} (requires sudo)..."
        sudo mv "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    fi

    chmod +x "${INSTALL_DIR}/${BINARY}"
    echo "${BINARY} ${latest} installed to ${INSTALL_DIR}/${BINARY}"
}

main
