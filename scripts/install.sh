#!/usr/bin/env bash
set -euo pipefail

# One-machine install: build hy2mgr and run `hy2mgr install`.
# Usage (recommended):
#   curl -fsSL https://raw.githubusercontent.com/<YOU>/hy2mgr/main/scripts/install.sh | sudo bash
# Optional:
#   HY2MGR_REPO=https://github.com/<YOU>/hy2mgr.git  (default: current repo URL)
#   HY2MGR_REF=main

HY2MGR_REPO="${HY2MGR_REPO:-https://github.com/yuzeguitarist/hy2mgr.git}"
HY2MGR_REF="${HY2MGR_REF:-main}"

color_enabled() {
  [[ -t 1 ]] && [[ -z "${NO_COLOR:-}" ]] && [[ "${TERM:-}" != "dumb" ]]
}

cprint() {
  local code="$1"
  shift
  if color_enabled; then
    printf "\033[%sm%s\033[0m" "$code" "$*"
  else
    printf "%s" "$*"
  fi
}

info() { cprint "1;34" "$1"; echo; }
ok() { cprint "1;32" "$1"; echo; }
warn() { cprint "1;33" "$1"; echo; }

detect_ip() {
  local ip=""
  ip="$(curl -fsSL https://api.ipify.org 2>/dev/null || true)"
  if [[ -z "$ip" ]]; then
    ip="$(hostname -I 2>/dev/null | awk '{print $1}')"
  fi
  if [[ -z "$ip" ]]; then
    ip="YOUR_VPS_IP"
  fi
  echo "$ip"
}

if [[ $EUID -ne 0 ]]; then
  warn "Run as root (sudo)."
  exit 1
fi

info "==> Installing dependencies"
export DEBIAN_FRONTEND=noninteractive
apt-get update -y
apt-get install -y --no-install-recommends ca-certificates curl git build-essential golang

workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT
cd "$workdir"

info "==> Cloning $HY2MGR_REPO ($HY2MGR_REF)"
git clone --depth 1 --branch "$HY2MGR_REF" "$HY2MGR_REPO" hy2mgr
cd hy2mgr

info "==> Building"
go mod tidy
go build -o /usr/local/bin/hy2mgr ./main.go

info "==> Running hy2mgr install"
hy2mgr install

ok "==> Done."
ip="$(detect_ip)"
echo "Open Web UI:"
echo "  http://${ip}:3333 (ensure firewall allows TCP/3333)"
echo "Or via SSH port-forward (safer):"
echo "  ssh -L 3333:127.0.0.1:3333 root@${ip}"
echo "  then open http://127.0.0.1:3333"
