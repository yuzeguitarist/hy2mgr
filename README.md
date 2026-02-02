# HY2 管理套件（Hysteria2 Manager / hy2mgr）

目标：在一台 Linux VPS 上**低成本**、**安全默认**地安装/配置/管理 Hysteria2 服务端，并提供**节点管理**（URI/二维码/订阅）。

- **后端/CLI：Go 单二进制**
- **Web UI：嵌入式静态页面（无需额外构建）**
- 支持：Ubuntu/Debian（systemd）

> 说明：本项目默认自签证书，不依赖域名；默认 UDP/443，冲突自动切换端口；默认 SNI：`www.bing.com`；默认 masquerade：`proxy + rewriteHost`（可配置）。官方配置 schema 与安装方式参考 Hysteria2 文档与官方安装脚本。 citeturn2view1turn4view0turn7view0

---

## 一条命令“单机部署”（推荐）

在全新 VPS 上：

```bash
curl -fsSL https://raw.githubusercontent.com/yuzeguitarist/hy2mgr/main/scripts/install.sh | sudo bash
```

完成后输出：
- 管理员用户名/密码（只显示一次）
- 订阅 token 与订阅 URL（只显示一次）
- Web UI 默认监听 `0.0.0.0:3333`

### 访问 Web UI
直接访问：
http://YOUR_VPS_IP:3333

或通过 SSH 端口转发（更安全）：
```bash
ssh -L 3333:127.0.0.1:3333 root@YOUR_VPS_IP
# 浏览器打开 http://127.0.0.1:3333
```

---

## 手动部署（审计友好）

见：`scripts/manual-deploy.md`

---

## 常用 CLI

> 需要改系统配置/服务的命令必须 root 运行。

### 服务端管理
```bash
sudo hy2mgr install
sudo hy2mgr apply --dry-run
sudo hy2mgr status
sudo hy2mgr logs --lines 200
sudo hy2mgr restore --backup /etc/hysteria/config.yaml.<timestamp>.bak
sudo hy2mgr uninstall --purge
```

### 节点管理
```bash
sudo hy2mgr node add --name my-phone
sudo hy2mgr node ls
sudo hy2mgr node disable --id <ID>
sudo hy2mgr node enable  --id <ID>
sudo hy2mgr node reset   --id <ID>
sudo hy2mgr node rm      --id <ID>
```

### 导出
```bash
hy2mgr export uri --id <ID>
hy2mgr export qrcode --id <ID> --out ./node.png
sudo hy2mgr export subscription --rotate
```

### 证书
```bash
sudo hy2mgr cert fingerprint
sudo hy2mgr cert rotate
```

---

## 目录与关键文件

- Hysteria 配置：`/etc/hysteria/config.yaml`
- TLS 证书：`/etc/hysteria/cert.crt`
- TLS 私钥：`/etc/hysteria/cert.key`
- hy2mgr 状态：`/etc/hy2mgr/state.json`（0600，root-only）
- 审计日志：`/var/log/hy2mgr/audit.log`（jsonl）

---

## 故障排查（至少 5 项）

1) **连不上/超时**
- 先确认服务是否运行：`sudo hy2mgr status`
- 看日志是否报错：`sudo hy2mgr logs --lines 200`
- 确认端口：Dashboard 或 `hy2mgr status`

2) **UDP 端口未放行（最常见）**
- 本机防火墙：hy2mgr 会尝试放行 UFW/firewalld/iptables
- **云厂商安全组/防火墙**：必须在控制台放行 **UDP/<端口>**（默认 443；冲突会自动换端口，务必按实际端口放行）

3) **UDP/443 被占用**
- hy2mgr 会自动探测并切换到候选端口（例如 8443/2053/2083/2096 等）
- 用 `hy2mgr status` 确认最终端口并放行 UDP

4) **`tls.key permission denied`**
- 常见原因：Hysteria2 服务用非 root 用户运行（如 `User=hysteria`），但 `cert.key` 只有 root 可读
- `hy2mgr apply` 会自动修复：将 `/etc/hysteria` 目录与 key 文件设置为 root:<service-group> 且 key 为 0640
- 修复后重启服务：`sudo systemctl restart hysteria-server.service`

5) **配置语法问题**
- hy2mgr 写入前会做 YAML 校验（结构化解析 + 必填字段）
- 若你手动改了 `/etc/hysteria/config.yaml` 导致无法启动，可使用：
  - `ls /etc/hysteria/config.yaml.*.bak`
  - `sudo hy2mgr restore --backup <某个.bak>`

6) **证书 pinSHA256 变化导致客户端不通**
- `hy2mgr cert rotate` 会生成新自签证书，pin 会变化
- 重新导出节点 URI 或订阅更新即可

---

## 备份与恢复

### 备份
```bash
sudo tar czf hy2mgr-backup.tgz /etc/hy2mgr /etc/hysteria /var/log/hy2mgr
```

### 恢复
```bash
sudo tar xzf hy2mgr-backup.tgz -C /
sudo hy2mgr apply
```

---

## 升级

```bash
# 重新拉取代码并替换 /usr/local/bin/hy2mgr
sudo hy2mgr install
```

`install` 会走**幂等**流程：存在则升级，不会反复破坏系统。

---

## 卸载

```bash
sudo hy2mgr uninstall
# 如果你想删掉所有状态与日志：
sudo hy2mgr uninstall --purge
```

---

## 安全提示（重要）
- Web UI 默认 **监听 0.0.0.0:3333**。请在云安全组/防火墙放行 TCP/3333；公网访问建议优先使用 **HTTPS 反代**。
- 如需更安全的访问方式，可使用 **SSH 端口转发**。
- 如果必须公网访问，请用 Nginx/Caddy 做 HTTPS 反代，并加额外访问控制（IP allowlist / 2FA）。
- 订阅 URL 必须保密：token 一旦泄露，任何人都能获取节点链接；可用 `hy2mgr export subscription --rotate` 立即吊销旧 token。
