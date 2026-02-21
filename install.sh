#!/bin/sh
set -e

OWNER="nikmd1306"
REPO="cwai"
BINARY="cwai"
GITHUB="https://github.com"
INSTALL_DIR="/usr/local/bin"

usage() {
    cat <<EOF
Usage: install.sh [-b install_dir]

Flags:
    -b    installation directory (default: ${INSTALL_DIR})
EOF
    exit 2
}

uname_os() {
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$os" in
        msys* | mingw* | cygwin*) os="windows" ;;
    esac
    printf '%s' "$os"
}

uname_arch() {
    arch=$(uname -m)
    case "$arch" in
        x86_64)  arch="amd64" ;;
        aarch64) arch="arm64" ;;
        arm64)   arch="arm64" ;;
    esac
    printf '%s' "$arch"
}

http_download() {
    url=$1
    dest=$2
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL -o "$dest" "$url"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "$dest" "$url"
    else
        echo "error: curl or wget required" >&2
        return 1
    fi
}

http_get_redirect() {
    url=$1
    if command -v curl >/dev/null 2>&1; then
        curl -fsSI -o /dev/null -w '%{url_effective}' "$url" 2>/dev/null
    elif command -v wget >/dev/null 2>&1; then
        wget -S --spider --max-redirect=0 "$url" 2>&1 | grep -i 'Location:' | sed 's/.*Location: *//' | tr -d '\r'
    else
        echo "error: curl or wget required" >&2
        return 1
    fi
}

hash_sha256() {
    file=$1
    if command -v gsha256sum >/dev/null 2>&1; then
        gsha256sum "$file" | cut -d ' ' -f 1
    elif command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$file" | cut -d ' ' -f 1
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$file" | cut -d ' ' -f 1
    elif command -v openssl >/dev/null 2>&1; then
        openssl dgst -sha256 "$file" | sed 's/^.* //'
    else
        echo "error: no sha256 tool found" >&2
        return 1
    fi
}

main() {
    while getopts "b:h" opt; do
        case "$opt" in
            b) INSTALL_DIR="$OPTARG" ;;
            h) usage ;;
            *) usage ;;
        esac
    done

    os=$(uname_os)
    arch=$(uname_arch)
    echo "Detected platform: ${os}/${arch}"

    redirect_url=$(http_get_redirect "${GITHUB}/${OWNER}/${REPO}/releases/latest")
    tag=$(printf '%s' "$redirect_url" | grep -oE '[^/]+$')
    if [ -z "$tag" ]; then
        echo "error: unable to detect latest version" >&2
        exit 1
    fi
    version="${tag#v}"
    echo "Latest version: ${tag}"

    ext="tar.gz"
    if [ "$os" = "windows" ]; then
        ext="zip"
    fi
    archive="${BINARY}_${version}_${os}_${arch}.${ext}"
    checksums="${BINARY}_${version}_checksums.txt"

    tmpdir=$(mktemp -d)
    trap 'rm -rf "$tmpdir"' EXIT

    base_url="${GITHUB}/${OWNER}/${REPO}/releases/download/${tag}"
    echo "Downloading ${archive}..."
    http_download "${base_url}/${archive}" "${tmpdir}/${archive}"
    http_download "${base_url}/${checksums}" "${tmpdir}/${checksums}"

    echo "Verifying checksum..."
    expected=$(grep "${archive}" "${tmpdir}/${checksums}" | cut -d ' ' -f 1)
    if [ -z "$expected" ]; then
        echo "error: archive not found in checksums file" >&2
        exit 1
    fi
    actual=$(hash_sha256 "${tmpdir}/${archive}")
    if [ "$expected" != "$actual" ]; then
        echo "error: checksum mismatch" >&2
        echo "  expected: ${expected}" >&2
        echo "  actual:   ${actual}" >&2
        exit 1
    fi
    echo "Checksum verified."

    echo "Extracting..."
    if [ "$ext" = "zip" ]; then
        unzip -o -q "${tmpdir}/${archive}" -d "${tmpdir}"
    else
        tar -xzf "${tmpdir}/${archive}" -C "${tmpdir}"
    fi

    mkdir -p "${INSTALL_DIR}"
    if [ -w "${INSTALL_DIR}" ]; then
        install -m 755 "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    else
        echo "Elevated permissions required to install to ${INSTALL_DIR}"
        sudo install -m 755 "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    fi

    echo "${BINARY} installed to ${INSTALL_DIR}/${BINARY}"
}

main "$@"
