#!/usr/bin/env bash
set -euo pipefail

REPO_SLUG="${PASSKEY_SUDO_REPO:-hamdyelbatal122/sudo-passkey}"
BIN_NAME="passkey-sudo"
INSTALL_DIR="${PASSKEY_SUDO_INSTALL_DIR:-/usr/local/bin}"
WORKDIR="${TMPDIR:-/tmp}/passkey-sudo-install"
GO_VERSION="${PASSKEY_SUDO_GO_VERSION:-1.23.10}"

detect_os() {
  case "$(uname -s)" in
    Linux) echo "linux" ;;
    Darwin) echo "darwin" ;;
    *) echo "unsupported" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) echo "unsupported" ;;
  esac
}

install_binary() {
  local source_bin="$1"
  if [[ -w "$INSTALL_DIR" ]]; then
    install -m 0755 "$source_bin" "$INSTALL_DIR/$BIN_NAME"
  elif command -v sudo >/dev/null 2>&1; then
    sudo install -m 0755 "$source_bin" "$INSTALL_DIR/$BIN_NAME"
  else
    echo "Need write permission to $INSTALL_DIR. Re-run with sudo or set PASSKEY_SUDO_INSTALL_DIR."
    exit 1
  fi
}

ensure_go() {
  if command -v go >/dev/null 2>&1; then
    return 0
  fi

  local os arch go_tgz url go_root
  os="$(detect_os)"
  arch="$(detect_arch)"

  if [[ "$os" == "unsupported" || "$arch" == "unsupported" ]]; then
    echo "Go is not installed and this platform is not supported for auto-install."
    echo "Please install Go 1.23+ manually: https://go.dev/doc/install"
    exit 1
  fi

  echo "Go not found. Installing Go ${GO_VERSION} (${os}/${arch})..."
  go_tgz="$WORKDIR/go.tgz"
  url="https://go.dev/dl/go${GO_VERSION}.${os}-${arch}.tar.gz"
  curl -fsSL "$url" -o "$go_tgz"

  if [[ -w /usr/local ]]; then
    rm -rf /usr/local/go
    tar -C /usr/local -xzf "$go_tgz"
    go_root="/usr/local/go"
  elif command -v sudo >/dev/null 2>&1; then
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "$go_tgz"
    go_root="/usr/local/go"
  else
    go_root="$HOME/.local/go"
    rm -rf "$go_root"
    mkdir -p "$HOME/.local"
    tar -C "$HOME/.local" -xzf "$go_tgz"
  fi

  export PATH="$go_root/bin:$PATH"

  if ! command -v go >/dev/null 2>&1; then
    echo "Failed to make Go available in PATH."
    exit 1
  fi

  echo "Go installed: $(go version)"
}

latest_tag() {
  curl -fsSL "https://api.github.com/repos/${REPO_SLUG}/releases/latest" \
    | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' \
    | head -n 1
}

download_release_binary() {
  local tag="$1"
  local os arch version base asset candidate out
  os="$(detect_os)"
  arch="$(detect_arch)"

  if [[ "$os" == "unsupported" || "$arch" == "unsupported" ]]; then
    return 1
  fi

  version="${tag#v}"
  base="https://github.com/${REPO_SLUG}/releases/download/${tag}"

  for candidate in \
    "${BIN_NAME}_${version}_${os}_${arch}.tar.gz" \
    "${BIN_NAME}_${tag}_${os}_${arch}.tar.gz"
  do
    out="$WORKDIR/$candidate"
    if curl -fsSL "$base/$candidate" -o "$out"; then
      tar -xzf "$out" -C "$WORKDIR"
      if [[ -f "$WORKDIR/$BIN_NAME" ]]; then
        install_binary "$WORKDIR/$BIN_NAME"
        return 0
      fi
    fi
  done

  return 1
}

build_from_source() {
  local tag="$1"
  local source_tgz source_dir

  ensure_go

  source_tgz="$WORKDIR/source.tgz"
  curl -fsSL "https://github.com/${REPO_SLUG}/archive/refs/tags/${tag}.tar.gz" -o "$source_tgz"
  tar -xzf "$source_tgz" -C "$WORKDIR"

  source_dir="$WORKDIR/sudo-passkey-${tag#v}"
  if [[ ! -d "$source_dir" ]]; then
    source_dir="$WORKDIR/sudo-passkey-$tag"
  fi
  if [[ ! -d "$source_dir" ]]; then
    source_dir="$(find "$WORKDIR" -maxdepth 1 -type d -name 'sudo-passkey-*' | head -n 1)"
  fi

  if [[ -z "$source_dir" || ! -d "$source_dir" ]]; then
    echo "Failed to locate extracted source directory."
    exit 1
  fi

  (cd "$source_dir" && go build -o "$WORKDIR/$BIN_NAME" ./cmd/passkey-sudo)
  install_binary "$WORKDIR/$BIN_NAME"
}

rm -rf "$WORKDIR"
mkdir -p "$WORKDIR"

TAG="$(latest_tag || true)"
if [[ -z "$TAG" ]]; then
  echo "Could not fetch latest release tag from GitHub API."
  echo "Set PASSKEY_SUDO_REPO or check network/GitHub access."
  exit 1
fi

echo "Installing ${BIN_NAME} from latest release: ${TAG}"

if download_release_binary "$TAG"; then
  echo "Installed ${BIN_NAME} binary from release assets."
else
  echo "Release asset not available for this platform. Falling back to source build."
  build_from_source "$TAG"
  echo "Installed ${BIN_NAME} by building from source."
fi

echo "Installed ${BIN_NAME} to ${INSTALL_DIR}/${BIN_NAME}"
echo "Next: passkey-sudo init && passkey-sudo enroll"
