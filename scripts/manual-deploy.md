# Manual deploy (audit-friendly)

## 0. Preconditions
- Ubuntu/Debian with systemd
- root (or sudo)

## 1) Install deps
```bash
sudo apt-get update -y
sudo apt-get install -y --no-install-recommends ca-certificates curl git build-essential golang
```

## 2) Build hy2mgr
```bash
git clone https://github.com/<YOU>/hy2mgr.git
cd hy2mgr
go mod tidy
go test ./... -count=1
go build -o ./hy2mgr ./main.go
sudo install -m 0755 ./hy2mgr /usr/local/bin/hy2mgr
```

## 3) Install & configure hysteria2 + hy2mgr
```bash
sudo /usr/local/bin/hy2mgr install
```
This step:
- installs/updates hysteria2 with the official script (get.hy2.sh) citeturn7view0
- generates self-signed TLS cert/key under `/etc/hysteria/`
- writes `/etc/hysteria/config.yaml` per official schema citeturn2view1turn4view0
- enables and restarts systemd services

## 4) Validate
```bash
sudo hy2mgr status
sudo hy2mgr node ls
sudo hy2mgr export uri --id <NODE_ID>
```

## 5) Access Web UI (recommended: SSH port-forward)
```bash
ssh -L 3333:127.0.0.1:3333 root@YOUR_VPS_IP
# then open http://127.0.0.1:3333
```
