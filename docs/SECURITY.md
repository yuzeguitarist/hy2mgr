# 安全策略与威胁模型（简版）

## 资产（Assets）
- Hysteria2 服务端私钥：`/etc/hysteria/cert.key`
- 节点密码（userpass auth）
- 管理员口令（bcrypt 哈希存储）
- 订阅 token（只存 SHA-256，明文只显示一次）
- 审计日志（用于追踪变更）

## 信任边界
- VPS 本机（root 用户）是最高信任边界。
- Web UI 默认绑定 `0.0.0.0:3333`，需视为不可信网络，建议强制 HTTPS + 反代安全措施。
- 如需更高安全性，可改为仅本地监听并通过 SSH port-forward 访问。

## 威胁与对策

### 1) 管理口暴露被爆破
**对策**
- 默认监听 0.0.0.0:3333，必须配合防火墙/反代安全措施。
- 登录必须用户名/密码；管理员密码 bcrypt 哈希存储。
- 可选 2FA（TOTP）——本项目代码已预留（在 state 中有字段），建议在生产环境启用并配合反代限制。
- 建议反代层加 IP allowlist / basic auth / fail2ban。

### 2) 订阅 URL 被爬虫扫描
**对策**
- 订阅 URL 携带随机 token（高熵），并且 token 在服务端只存 hash，泄露风险降低。
- 支持一键旋转 token（旧 token 立即失效）。

### 3) 私钥/敏感信息泄露到日志
**对策**
- 审计日志只记录“动作/对象”，不记录节点密码、管理员密码、token 明文。
- CLI 输出 token/管理员初始密码仅在首次 install 时显示一次（用户需自行保存）。

### 4) `tls.key permission denied` 造成服务不可用
**对策**
- 自动修复目录与 key 文件权限：`root:<service-group>` + `0640`，避免服务用户读取失败。

### 5) 配置变更导致服务不可用
**对策**
- 写入前做 YAML 结构校验
- 写入时自动备份 `/etc/hysteria/config.yaml.<timestamp>.bak`
- 提供 `hy2mgr restore` 一键回滚

## 默认安全配置
- Web UI：`0.0.0.0:3333`
- 状态文件：`/etc/hy2mgr/state.json` (0600)
- Hysteria 目录：`/etc/hysteria` (0750, root:<svc-group>)
- key：0640；cert：0644；config：0640

## 审计
所有以下动作写入 `/var/log/hy2mgr/audit.log`（jsonl）：
- 节点增删/启用禁用/重置密码
- 设置变更（端口/SNI/masquerade）
- 订阅 token 旋转
- 证书轮换
