package cmd

import (
	"fmt"

	"github.com/yuzeguitarist/hy2mgr/internal/app"
	"github.com/yuzeguitarist/hy2mgr/internal/service"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Idempotently apply cert, config, permissions, firewall rules, and restart hysteria2",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := app.MustBeRoot(); err != nil {
			return err
		}
		dry, _ := cmd.Flags().GetBool("dry-run")
		st := mustLoadState()
		if dry {
			fmt.Println("==> Preview of config (passwords masked)")
			prev, _ := service.SaveConfigPreview(st)
			fmt.Println(prev)
		}
		if err := service.Apply(st, dry); err != nil {
			return err
		}
		fmt.Println("Applied.")
		return nil
	},
}

func init() {
	applyCmd.Flags().Bool("dry-run", false, "preview changes without applying")
}
