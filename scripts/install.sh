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

if [[ $EUID -ne 0 ]]; then
  echo "Run as root (sudo)."
  exit 1
fi

echo "==> Installing dependencies"
export DEBIAN_FRONTEND=noninteractive
apt-get update -y
apt-get install -y --no-install-recommends ca-certificates curl git build-essential golang

workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT
cd "$workdir"

echo "==> Cloning $HY2MGR_REPO ($HY2MGR_REF)"
git clone --depth 1 --branch "$HY2MGR_REF" "$HY2MGR_REPO" hy2mgr
cd hy2mgr

echo "==> Building"
go mod tidy
go build -o /usr/local/bin/hy2mgr ./main.go

echo "==> Running hy2mgr install"
hy2mgr install

echo "==> Done."
echo "Open Web UI:"
echo "  http://YOUR_VPS_IP:3333 (ensure firewall allows TCP/3333)"
echo "Or via SSH port-forward (safer):"
echo "  ssh -L 3333:127.0.0.1:3333 root@YOUR_VPS_IP"
echo "  then open http://127.0.0.1:3333"
