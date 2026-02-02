package app

const (
	// Hysteria
	HysteriaConfigPath = "/etc/hysteria/config.yaml"
	HysteriaCertPath   = "/etc/hysteria/cert.crt"
	HysteriaKeyPath    = "/etc/hysteria/cert.key"
	HysteriaService    = "hysteria-server.service"

	// Manager state
	StateDir      = "/etc/hy2mgr"
	StatePath     = "/etc/hy2mgr/state.json"
	StateBackups  = "/etc/hy2mgr/backups"

	// Manager audit log
	AuditDir  = "/var/log/hy2mgr"
	AuditPath = "/var/log/hy2mgr/audit.log"

	// Manager systemd
	ManagerService = "hy2mgr.service"
)
