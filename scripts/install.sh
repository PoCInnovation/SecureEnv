#!/bin/sh
# Install the secureenv CLI from a GitHub release.
#
#   curl -fsSL https://raw.githubusercontent.com/PoCInnovation/SecureEnv/main/scripts/install.sh | sh
#
# Environment:
#   SECUREENV_VERSION      version to install, e.g. v1.2.0 (default: latest)
#   SECUREENV_INSTALL_DIR  target directory (default: /usr/local/bin if
#                          writable, otherwise ~/.local/bin)
#   SECUREENV_BASE_URL     download base URL, for mirrors (default: GitHub)
set -eu

repo="PoCInnovation/SecureEnv"
base_url="${SECUREENV_BASE_URL:-https://github.com/$repo/releases}"

fail() {
  echo "secureenv install: $*" >&2
  exit 1
}

need() {
  command -v "$1" > /dev/null 2>&1 || fail "$1 is required"
}

need curl
need tar
need uname

case "$(uname -s)" in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  *) fail "unsupported OS $(uname -s); on Windows download the zip from $base_url" ;;
esac

case "$(uname -m)" in
  x86_64 | amd64) arch=amd64 ;;
  arm64 | aarch64) arch=arm64 ;;
  *) fail "unsupported architecture $(uname -m)" ;;
esac

version="${SECUREENV_VERSION:-latest}"
if [ "$version" = "latest" ]; then
  latest_url=$(curl -fsSLI -o /dev/null -w '%{url_effective}' "$base_url/latest") ||
    fail "cannot resolve the latest release"
  version="${latest_url##*/}"
fi
case "$version" in
  v*) ;;
  *) version="v$version" ;;
esac

archive="secureenv_${version#v}_${os}_${arch}.tar.gz"
download_url="$base_url/download/$version"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "Downloading secureenv $version for $os/$arch..." >&2
curl -fsSLo "$tmp/$archive" "$download_url/$archive" || fail "cannot download $download_url/$archive"
curl -fsSLo "$tmp/checksums.txt" "$download_url/checksums.txt" || fail "cannot download checksums.txt"

expected=$(awk -v file="$archive" '$2 == file { print $1 }' "$tmp/checksums.txt")
[ -n "$expected" ] || fail "no checksum for $archive"
if command -v sha256sum > /dev/null 2>&1; then
  actual=$(sha256sum "$tmp/$archive" | awk '{ print $1 }')
else
  need shasum
  actual=$(shasum -a 256 "$tmp/$archive" | awk '{ print $1 }')
fi
[ "$expected" = "$actual" ] || fail "checksum mismatch for $archive"

tar -xzf "$tmp/$archive" -C "$tmp" secureenv

install_dir="${SECUREENV_INSTALL_DIR:-}"
if [ -z "$install_dir" ]; then
  if [ -w /usr/local/bin ]; then
    install_dir=/usr/local/bin
  else
    install_dir="$HOME/.local/bin"
  fi
fi
mkdir -p "$install_dir"
install -m 0755 "$tmp/secureenv" "$install_dir/secureenv"

echo "Installed $("$install_dir/secureenv" version) to $install_dir/secureenv" >&2
case ":$PATH:" in
  *":$install_dir:"*) ;;
  *) echo "Add $install_dir to your PATH to use secureenv." >&2 ;;
esac
