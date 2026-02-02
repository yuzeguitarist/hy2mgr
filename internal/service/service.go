package service

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/yuzeguitarist/hy2mgr/internal/app"
	"github.com/yuzeguitarist/hy2mgr/internal/crypto"
	"github.com/yuzeguitarist/hy2mgr/internal/firewall"
	"github.com/yuzeguitarist/hy2mgr/internal/hysteria"
	"github.com/yuzeguitarist/hy2mgr/internal/netutil"
	"github.com/yuzeguitarist/hy2mgr/internal/state"
	"github.com/yuzeguitarist/hy2mgr/internal/systemd"
)


type PermSpec struct {
	Path string
	Mode os.FileMode
}

func DesiredPerms() []PermSpec {
	return []PermSpec{
		{Path: filepath.Dir(app.HysteriaKeyPath), Mode: 0750},
		{Path: app.HysteriaKeyPath, Mode: 0640},
		{Path: app.HysteriaCertPath, Mode: 0644},
		{Path: app.HysteriaConfigPath, Mode: 0640},
		{Path: filepath.Dir(app.HysteriaConfigPath), Mode: 0750},
	}
}

// Apply is the idempotent "desired state" reconciler.
func Apply(st *state.State, dryRun bool) error {
	// 1) ensure cert exists
	if _, err := os.Stat(app.HysteriaCertPath); err != nil {
		if err := RotateCert(st, dryRun); err != nil {
			return err
		}
	}

	// 2) ensure at least one node for auth
	if len(st.Nodes) == 0 {
		_, err := NodeAdd(st, "default", "", "")
		if err != nil {
			return err
		}
	}

	// 3) choose port if busy (prefer 443) per requirements
	candidates := []int{443, 8443, 2053, 2083, 2087, 2096, 10443}
	st.Settings.ListenPort = netutil.ChoosePort(st.Settings.ListenPort, candidates)

	// 4) build userpass map (enabled only)
	users := map[string]string{}
	for _, n := range st.Nodes {
		if n.Enabled {
			users[n.Username] = n.Password
		}
	}
	y, err := hysteria.GenerateYAML(st.Settings.ListenPort, app.HysteriaCertPath, app.HysteriaKeyPath, users, st.Settings.MasqueradeURL, st.Settings.MasqueradeRewrite)
	if err != nil {
		return err
	}
	if err := hysteria.ValidateYAML(y); err != nil {
		return err
	}

	// 5) backup & write /etc/hysteria/config.yaml with rollback
	if !dryRun {
		if err := app.EnsureDir(filepath.Dir(app.HysteriaConfigPath), 0750); err != nil {
			return err
		}
		if _, err := os.Stat(app.HysteriaConfigPath); err == nil {
			backup := app.HysteriaConfigPath + "." + app.NowRFC3339() + ".bak"
			_ = app.CopyFile(app.HysteriaConfigPath, backup, 0644)
		}
		if err := app.AtomicWriteFile(app.HysteriaConfigPath, 0640, y); err != nil {
			return err
		}
	}

	// 6) permission self-heal
	if err := FixKeyPermission(dryRun); err != nil {
		return err
	}

	// 7) firewall open UDP port
	_, _, _ = firewall.EnsureUDPPortOpen(st.Settings.ListenPort, dryRun)

	// 8) restart hysteria
	if !dryRun {
		_ = systemd.EnableNow(app.HysteriaService)
		_ = systemd.Restart(app.HysteriaService)
	}
	return nil
}

// RotateCert creates a new self-signed cert/key pair and writes to disk.
func RotateCert(st *state.State, dryRun bool) error {
	var ipAddrs []net.IP
	if st != nil && st.Settings.ListenHost != "" {
		if ip := net.ParseIP(st.Settings.ListenHost); ip != nil {
			ipAddrs = append(ipAddrs, ip)
		}
	}
	if len(ipAddrs) == 0 {
		if ip := net.ParseIP(detectPublicIP()); ip != nil {
			ipAddrs = append(ipAddrs, ip)
		}
	}
	if len(ipAddrs) == 0 {
		ipAddrs = append(ipAddrs, net.ParseIP("127.0.0.1"))
	}

	certPEM, keyPEM, _, err := crypto.GenerateSelfSigned(ipAddrs, 3650)
	if err != nil {
		return err
	}
	if dryRun {
		return nil
	}
	return crypto.WriteCertFiles(certPEM, keyPEM, 0, 0)
}

// FixKeyPermission repairs tls.key permission denied when hysteria runs as non-root.
func FixKeyPermission(dryRun bool) error {
	unit, _ := systemd.Cat(app.HysteriaService)
	svcUser := "hysteria"
	for _, l := range strings.Split(unit, "\n") {
		if strings.HasPrefix(strings.TrimSpace(l), "User=") {
			svcUser = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(l), "User="))
		}
	}
	u, err := user.Lookup(svcUser)
	if err != nil {
		return nil
	}
	gid := atoi(u.Gid)

	return applyPerms(gid, dryRun)
}


func applyPerms(gid int, dryRun bool) error {
	if dryRun {
		return nil
	}
	for _, p := range DesiredPerms() {
		_ = os.Chmod(p.Path, p.Mode)
		_ = os.Chown(p.Path, 0, gid) // best-effort; may fail in non-root contexts
	}
	return nil
}

func NodeAdd(st *state.State, name, username, password string) (*state.Node, error) {
	id, _ := app.RandToken(8)
	if username == "" {
		username = "u" + id
	}
	if password == "" {
		password, _ = app.RandToken(16)
	}
	n := state.Node{
		ID:        id,
		Name:      name,
		Username:  username,
		Password:  password,
		Enabled:   true,
		CreatedAt: app.NowRFC3339(),
		UpdatedAt: app.NowRFC3339(),
	}
	st.Nodes = append(st.Nodes, n)
	if err := Apply(st, false); err != nil {
		return nil, err
	}
	_ = st.SaveAtomic()
	return &n, nil
}

func NodeDelete(st *state.State, id string) error {
	idx := -1
	for i := range st.Nodes {
		if st.Nodes[i].ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("node not found")
	}
	st.Nodes = append(st.Nodes[:idx], st.Nodes[idx+1:]...)
	if err := Apply(st, false); err != nil {
		return err
	}
	_ = st.SaveAtomic()
	return nil
}

func NodeSetEnabled(st *state.State, id string, enabled bool) error {
	for i := range st.Nodes {
		if st.Nodes[i].ID == id {
			st.Nodes[i].Enabled = enabled
			st.Nodes[i].UpdatedAt = app.NowRFC3339()
			if err := Apply(st, false); err != nil {
				return err
			}
			_ = st.SaveAtomic()
			return nil
		}
	}
	return fmt.Errorf("node not found")
}

func NodeResetPassword(st *state.State, id string) error {
	for i := range st.Nodes {
		if st.Nodes[i].ID == id {
			pass, _ := app.RandToken(16)
			st.Nodes[i].Password = pass
			st.Nodes[i].UpdatedAt = app.NowRFC3339()
			if err := Apply(st, false); err != nil {
				return err
			}
			_ = st.SaveAtomic()
			return nil
		}
	}
	return fmt.Errorf("node not found")
}

func NodeURI(st *state.State, id string) (string, error) {
	var n *state.Node
	for i := range st.Nodes {
		if st.Nodes[i].ID == id {
			n = &st.Nodes[i]
			break
		}
	}
	if n == nil {
		return "", fmt.Errorf("node not found")
	}
	pin, _ := crypto.ParseCertPin(app.HysteriaCertPath)
	host := st.Settings.ListenHost
	if host == "" {
		host = detectPublicIP()
	}
	auth := url.QueryEscape(n.Username + ":" + n.Password)
	q := url.Values{}
	q.Set("insecure", "1")
	q.Set("sni", st.Settings.SNI)
	if pin != "" {
		q.Set("pinSHA256", pin)
	}
	return fmt.Sprintf("hysteria2://%s@%s:%d/?%s", auth, host, st.Settings.ListenPort, q.Encode()), nil
}

func detectPublicIP() string {
	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		if (iface.Flags&net.FlagUp) == 0 || (iface.Flags&net.FlagLoopback) != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, a := range addrs {
			ip := parseIP(a.String())
			if ip == "" {
				continue
			}
			if strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "172.") || strings.HasPrefix(ip, "127.") {
				continue
			}
			return ip
		}
	}
	return "YOUR_VPS_IP"
}

func parseIP(s string) string {
	if strings.Contains(s, "/") {
		s = strings.SplitN(s, "/", 2)[0]
	}
	if net.ParseIP(s) != nil {
		return s
	}
	return ""
}

func SubscriptionRotate(st *state.State) (string, string, error) {
	token, err := app.RandToken(18)
	if err != nil {
		return "", "", err
	}
	sum := sha256.Sum256([]byte(token))
	st.Subscription.TokenSHA256 = hex.EncodeToString(sum[:])
	st.Subscription.CreatedAt = app.NowRFC3339()
	st.Subscription.RevokedAt = ""
	_ = st.SaveAtomic()
	return token, "/sub/" + token, nil
}

func SubscriptionVerify(st *state.State, token string) bool {
	if st.Subscription.TokenSHA256 == "" || st.Subscription.RevokedAt != "" {
		return false
	}
	sum := sha256.Sum256([]byte(token))
	a := hex.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(a), []byte(st.Subscription.TokenSHA256)) == 1
}

func SaveConfigPreview(st *state.State) (string, error) {
	users := map[string]string{}
	for _, n := range st.Nodes {
		if n.Enabled {
			users[n.Username] = "***"
		}
	}
	y, err := hysteria.GenerateYAML(st.Settings.ListenPort, app.HysteriaCertPath, app.HysteriaKeyPath, users, st.Settings.MasqueradeURL, st.Settings.MasqueradeRewrite)
	if err != nil {
		return "", err
	}
	return string(y), nil
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}
