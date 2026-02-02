package cmd

import (
	"fmt"

	"github.com/yuzeguitarist/hy2mgr/internal/app"
	"github.com/yuzeguitarist/hy2mgr/internal/crypto"
	"github.com/yuzeguitarist/hy2mgr/internal/systemd"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of hysteria2 and hy2mgr",
	RunE: func(cmd *cobra.Command, args []string) error {
		st := mustLoadState()
		fmt.Println("HY2 Manager state:", app.StatePath)
		fmt.Println("Listen UDP port:", st.Settings.ListenPort)
		pin, _ := crypto.ParseCertPin(app.HysteriaCertPath)
		fmt.Println("Cert pinSHA256:", pin)

		fmt.Println("\n== hysteria-server.service ==")
		out, _ := systemd.Status(app.HysteriaService)
		fmt.Println(out)

		fmt.Println("\n== hy2mgr.service ==")
		out2, _ := systemd.Status(app.ManagerService)
		fmt.Println(out2)
		return nil
	},
}
