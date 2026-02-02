package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/example/hy2mgr/internal/app"
	"github.com/example/hy2mgr/internal/systemd"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall hysteria2 (official script --remove) and optionally remove hy2mgr state",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := app.MustBeRoot(); err != nil {
			return err
		}
		purge, _ := cmd.Flags().GetBool("purge")
		dry, _ := cmd.Flags().GetBool("dry-run")

		fmt.Println("==> Stopping services")
		if dry {
			fmt.Println("[dry-run] systemctl disable --now", app.ManagerService, app.HysteriaService)
		} else {
			_, _ = systemd.Systemctl("disable", "--now", app.ManagerService)
			_, _ = systemd.Systemctl("disable", "--now", app.HysteriaService)
		}

		fmt.Println("==> Removing Hysteria2 via official script")
		// Official script supports --remove citeturn7view0
		if dry {
			fmt.Println("[dry-run] bash <(curl -fsSL https://get.hy2.sh/) --remove")
		} else {
			c := exec.Command("bash", "-c", "bash <(curl -fsSL https://get.hy2.sh/) --remove")
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				return err
			}
		}

		fmt.Println("==> Removing hy2mgr unit")
		if !dry {
			_ = os.Remove("/etc/systemd/system/" + app.ManagerService)
			_, _ = systemd.Systemctl("daemon-reload")
		}

		if purge {
			fmt.Println("==> Purging state and logs (/etc/hy2mgr, /var/log/hy2mgr)")
			if !dry {
				_ = os.RemoveAll(app.StateDir)
				_ = os.RemoveAll(app.AuditDir)
			}
		} else {
			fmt.Println("State kept. Reinstall will reuse /etc/hy2mgr/state.json")
		}

		fmt.Println("Done.")
		return nil
	},
}

func init() {
	uninstallCmd.Flags().Bool("purge", false, "remove /etc/hy2mgr and /var/log/hy2mgr")
	uninstallCmd.Flags().Bool("dry-run", false, "preview changes without applying")
}
