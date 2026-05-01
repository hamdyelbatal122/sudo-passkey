#!/usr/bin/env bash
set -euo pipefail

REPO="${1:-https://github.com/hamdy/passkey-sudo.git}"
WORKDIR="${TMPDIR:-/tmp}/passkey-sudo-install"

rm -rf "$WORKDIR"
git clone "$REPO" "$WORKDIR"
cd "$WORKDIR"

if ! command -v go >/dev/null 2>&1; then
  echo "Go is required but not installed. Install Go 1.23+ and retry."
  exit 1
fi

make tidy
make build
sudo install -m 0755 bin/passkey-sudo /usr/local/bin/passkey-sudo

echo "Installed passkey-sudo to /usr/local/bin/passkey-sudo"
echo "Next: run 'passkey-sudo init && passkey-sudo enroll'"
