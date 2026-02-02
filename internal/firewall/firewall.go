package firewall

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/yuzeguitarist/hy2mgr/internal/app"
)

type Backend string

const (
	BackendUFW      Backend = "ufw"
	BackendFirewalld Backend = "firewalld"
	BackendIptables Backend = "iptables"
	BackendNone     Backend = "none"
)

func Detect() Backend {
	if app.CommandExists("ufw") {
		return BackendUFW
	}
	if app.CommandExists("firewall-cmd") {
		return BackendFirewalld
	}
	if app.CommandExists("iptables") {
		return BackendIptables
	}
	return BackendNone
}

func EnsureUDPPortOpen(port int, dryRun bool) (Backend, string, error) {
	b := Detect()
	switch b {
	case BackendUFW:
		msg, err := ensureUFW(port, dryRun)
		return b, msg, err
	case BackendFirewalld:
		msg, err := ensureFirewalld(port, dryRun)
		return b, msg, err
	case BackendIptables:
		msg, err := ensureIptables(port, dryRun)
		return b, msg, err
	default:
		return b, "No supported firewall backend detected; skipped local firewall rules.", nil
	}
}

func ensureUFW(port int, dryRun bool) (string, error) {
	out, _, _ := app.Exec("ufw", "status")
	if strings.Contains(out, "Status: inactive") {
		return "UFW detected but inactive; skipped.", nil
	}
	rule := fmt.Sprintf("%d/udp", port)
	if strings.Contains(out, rule) {
		return "UFW rule already present.", nil
	}
	if dryRun {
		return fmt.Sprintf("[dry-run] ufw allow %d/udp", port), nil
	}
	_, _, err := app.Exec("ufw", "allow", rule)
	return fmt.Sprintf("UFW allow %s added.", rule), err
}

func ensureFirewalld(port int, dryRun bool) (string, error) {
	out, _, _ := app.Exec("firewall-cmd", "--state")
	if strings.TrimSpace(out) != "running" {
		return "firewalld detected but not running; skipped.", nil
	}
	// list ports
	lp, _, _ := app.Exec("firewall-cmd", "--list-ports")
	re := regexp.MustCompile(`\b(\d+)/(tcp|udp)\b`)
	matches := re.FindAllStringSubmatch(lp, -1)
	for _, m := range matches {
		p, _ := strconv.Atoi(m[1])
		if p == port && m[2] == "udp" {
			return "firewalld port already open.", nil
		}
	}
	if dryRun {
		return fmt.Sprintf("[dry-run] firewall-cmd --permanent --add-port=%d/udp && firewall-cmd --reload", port), nil
	}
	_, _, err := app.Exec("firewall-cmd", "--permanent", "--add-port", fmt.Sprintf("%d/udp", port))
	if err != nil {
		return "", err
	}
	_, _, err = app.Exec("firewall-cmd", "--reload")
	return "firewalld rule added and reloaded.", err
}

func ensureIptables(port int, dryRun bool) (string, error) {
	// check existing rule
	out, _, _ := app.Exec("iptables", "-S", "INPUT")
	needle := fmt.Sprintf("-p udp -m udp --dport %d -j ACCEPT", port)
	if strings.Contains(out, needle) {
		return "iptables rule already present.", nil
	}
	if dryRun {
		return fmt.Sprintf("[dry-run] iptables -I INPUT -p udp --dport %d -j ACCEPT", port), nil
	}
	_, _, err := app.Exec("iptables", "-I", "INPUT", "-p", "udp", "--dport", fmt.Sprintf("%d", port), "-j", "ACCEPT")
	if err != nil {
		return "", err
	}
	return "iptables rule inserted. NOTE: persist rules yourself (e.g., iptables-persistent).", nil
}
