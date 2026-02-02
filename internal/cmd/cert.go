package cmd

import (
	"fmt"

	"github.com/yuzeguitarist/hy2mgr/internal/app"
	"github.com/yuzeguitarist/hy2mgr/internal/crypto"
	"github.com/yuzeguitarist/hy2mgr/internal/service"
	"github.com/spf13/cobra"
)

var certCmd = &cobra.Command{
	Use:   "cert",
	Short: "Manage self-signed cert",
}

var certRotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Rotate self-signed cert and restart hysteria2",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := app.MustBeRoot(); err != nil {
			return err
		}
		dry, _ := cmd.Flags().GetBool("dry-run")
		st := mustLoadState()
		if err := service.RotateCert(st, dry); err != nil {
			return err
		}
		if err := service.Apply(st, dry); err != nil {
			return err
		}
		fmt.Println("Rotated.")
		return nil
	},
}

var certFPcmd = &cobra.Command{
	Use:   "fingerprint",
	Short: "Print pinSHA256 of current cert",
	RunE: func(cmd *cobra.Command, args []string) error {
		pin, err := crypto.ParseCertPin(app.HysteriaCertPath)
		if err != nil {
			return err
		}
		fmt.Println(pin)
		return nil
	},
}

func init() {
	certCmd.AddCommand(certRotateCmd, certFPcmd)
	certRotateCmd.Flags().Bool("dry-run", false, "preview changes without applying")
}
