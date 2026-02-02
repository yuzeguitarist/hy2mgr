package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"

	"github.com/yuzeguitarist/hy2mgr/internal/app"
	"github.com/yuzeguitarist/hy2mgr/internal/service"
	"github.com/yuzeguitarist/hy2mgr/internal/systemd"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install/upgrade hysteria2 (official script), initialize state, write config, and enable services",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := app.MustBeRoot(); err != nil {
			return err
		}
		dry, _ := cmd.Flags().GetBool("dry-run")
		ver, _ := cmd.Flags().GetString("version")

		fmt.Println("==> Installing/Upgrading Hysteria2 via official script (get.hy2.sh)")
		// Official script usage: bash <(curl -fsSL https://get.hy2.sh/) citeturn7view0
		if dry {
			fmt.Println("[dry-run] bash <(curl -fsSL https://get.hy2.sh/)", versionArg(ver))
		} else {
			s := "bash <(curl -fsSL https://get.hy2.sh/) " + versionArg(ver)
			c := exec.Command("bash", "-c", s)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			c.Env = append(os.Environ(), "HYSTERIA_USER=hysteria")
			if err := c.Run(); err != nil {
				return fmt.Errorf("install hysteria2 failed: %w", err)
			}
		}

		st := mustLoadState()

		// admin bootstrap (only once)
		if st.Admin.PasswordBcrypt == "" {
			pw, _ := app.RandToken(12)
			h, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
			st.Admin.PasswordBcrypt = string(h)

			// subscription token
			token, _ := app.RandToken(18)
			sum := sha256.Sum256([]byte(token))
			st.Subscription.TokenSHA256 = hex.EncodeToString(sum[:])
			st.Subscription.CreatedAt = app.NowRFC3339()

			_ = st.SaveAtomic()

			fmt.Println("==> Admin credentials (shown once):")
			fmt.Println("    username:", st.Admin.Username)
			fmt.Println("    password:", pw)
			fmt.Println("==> Subscription URL token (shown once):")
			fmt.Println("    token:", token)
			fmt.Println("    url:   http://YOUR_VPS_IP:3333/sub/" + token)
			fmt.Println("    (Rotate later: hy2mgr export subscription --rotate)")
		}

		fmt.Println("==> Applying configuration (idempotent)")
		if err := service.Apply(st, dry); err != nil {
			return err
		}

		fmt.Println("==> Installing hy2mgr systemd service")
		if err := installManagerUnit(dry); err != nil {
			return err
		}
		if !dry {
			_ = systemd.EnableNow(app.ManagerService)
		}

		fmt.Println("Done.")
		fmt.Println("Web UI: http://YOUR_VPS_IP:3333 (ensure firewall allows TCP/3333)")
		return nil
	},
}

func versionArg(v string) string {
	if v == "" {
		return ""
	}
	return "--version " + v
}

func installManagerUnit(dry bool) error {
	unit := `[Unit]
Description=HY2 Manager (Hysteria2 Manager)
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/hy2mgr web --listen 0.0.0.0:3333
Restart=on-failure
RestartSec=2s
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/etc/hy2mgr /etc/hysteria /var/log/hy2mgr
RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX
LockPersonality=true
MemoryDenyWriteExecute=true

[Install]
WantedBy=multi-user.target
`
	path := "/etc/systemd/system/" + app.ManagerService
	if dry {
		fmt.Println("[dry-run] write", path)
		return nil
	}
	if err := os.WriteFile(path, []byte(unit), 0644); err != nil {
		return err
	}
	_, _ = systemd.Systemctl("daemon-reload")

	// install binary to /usr/local/bin/hy2mgr
	self, _ := os.Executable()
	dst := "/usr/local/bin/hy2mgr"
	if self != dst {
		return app.CopyFile(self, dst, 0755)
	}
	return nil
}

func init() {
	installCmd.Flags().Bool("dry-run", false, "preview changes without applying")
	installCmd.Flags().String("version", "", "install specified hysteria2 version (e.g., v2.7.0)")
}
