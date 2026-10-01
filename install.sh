#!/usr/bin/env sh
# Persona installer — no apt required.
# If run from inside a source checkout, it builds locally (no network).
# Otherwise it downloads a prebuilt binary from GitHub Releases, falling back
# to cloning + `go build`. Installs to ~/.local/bin (or /usr/local/bin with
# sudo) and adds it to PATH for your shell.
set -eu

REPO="Yescript/persona"
BINARY="persona"

info()  { printf '\033[36m==>\033[0m %s\n' "$1"; }
warn()  { printf '\033[33m!!\033[0m %s\n' "$1" >&2; }
die()   { printf '\033[31mxx\033[0m %s\n' "$1" >&2; exit 1; }

# 1. Detect OS + arch and map to a release asset name.
detect_target() {
  os="$(uname -s)"
  arch="$(uname -m)"
  case "$os" in
    Linux)  os="linux" ;;
    Darwin) os="darwin" ;;
    *) die "unsupported OS: $os (Linux and macOS only)" ;;
  esac
  case "$arch" in
    x86_64|amd64) arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
    *) die "unsupported arch: $arch" ;;
  esac
  TARGET="${os}_${arch}"
  OS="$os"
}

# 2. Pick an install directory.
pick_bindir() {
  if [ -w "/usr/local/bin" ] 2>/dev/null; then
    BINDIR="/usr/local/bin"
  else
    BINDIR="$HOME/.local/bin"
  fi
  mkdir -p "$BINDIR"
}

have() { command -v "$1" >/dev/null 2>&1; }

download() {
  url="$1"; dest="$2"
  if have curl; then
    curl -fsSL "$url" -o "$dest"
  elif have wget; then
    wget -qO "$dest" "$url"
  else
    return 1
  fi
}

# 3. Try prebuilt binary, else build from source.
install_binary() {
  base="https://github.com/${REPO}/releases/latest/download"
  asset="${BINARY}_${TARGET}"
  tmp="$(mktemp -d)"
  info "Downloading ${asset} ..."
  if download "${base}/${asset}.tar.gz" "${tmp}/a.tar.gz"; then
    tar -xzf "${tmp}/a.tar.gz" -C "${tmp}"
    bin="$(find "${tmp}" -type f -name "${BINARY}" | head -n1)"
    [ -n "$bin" ] || die "binary not found in archive"
    install -m 0755 "$bin" "${BINDIR}/${BINARY}"
    rm -rf "$tmp"
    return 0
  fi
  rm -rf "$tmp"
  return 1
}

# 0. If we are already inside a source checkout (the extracted zip or a git
#    clone), build straight from it — no network, no GitHub required.
build_local() {
  script_dir="$(cd "$(dirname "$0")" 2>/dev/null && pwd || true)"
  for d in "$script_dir" "$PWD"; do
    [ -n "$d" ] || continue
    if [ -f "$d/go.mod" ] && [ -d "$d/cmd/persona" ]; then
      have go || return 1
      info "Found local source in $d — building from it ..."
      ( cd "$d" && go build -ldflags "-s -w" -o "${BINDIR}/${BINARY}" ./cmd/persona )
      return 0
    fi
  done
  return 1
}

build_from_source() {
  have go || die "no prebuilt binary and Go is not installed — install Go from https://go.dev/dl/"
  info "Building from source with Go ..."
  tmp="$(mktemp -d)"
  if have git; then
    git clone --depth 1 "https://github.com/${REPO}.git" "${tmp}/src"
  else
    die "git not found; cannot clone source"
  fi
  ( cd "${tmp}/src" && go build -ldflags "-s -w" -o "${BINDIR}/${BINARY}" ./cmd/persona )
  rm -rf "$tmp"
}

# 4. Ensure BINDIR is on PATH for the user's shell.
ensure_path() {
  case ":$PATH:" in
    *":$BINDIR:"*) return 0 ;;
  esac
  shell_name="$(basename "${SHELL:-}")"
  case "$shell_name" in
    bash) rc="$HOME/.bashrc" ;;
    zsh)  rc="$HOME/.zshrc" ;;
    fish) rc="$HOME/.config/fish/config.fish" ;;
    *)    rc="$HOME/.profile" ;;
  esac
  mkdir -p "$(dirname "$rc")"
  if [ "$shell_name" = "fish" ]; then
    printf '\nset -gx PATH %s $PATH\n' "$BINDIR" >> "$rc"
  else
    printf '\nexport PATH="%s:$PATH"\n' "$BINDIR" >> "$rc"
  fi
  warn "Added $BINDIR to PATH in $rc — open a new shell or run: export PATH=\"$BINDIR:\$PATH\""
}

main() {
  detect_target
  pick_bindir
  info "Target: $TARGET  ·  install dir: $BINDIR"
  if build_local; then
    :
  elif install_binary; then
    :
  else
    warn "No local source and no prebuilt binary; cloning from GitHub."
    build_from_source
  fi
  ensure_path
  info "Verifying ..."
  if "${BINDIR}/${BINARY}" --version >/dev/null 2>&1; then
    "${BINDIR}/${BINARY}" --version
    info "Done. You can now run '${BINARY}' from anywhere."
  else
    die "installation completed but '${BINARY} --version' failed"
  fi
}

main "$@"
