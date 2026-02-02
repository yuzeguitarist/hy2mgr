package cmd

import (
	"fmt"

	"github.com/example/hy2mgr/internal/app"
	"github.com/example/hy2mgr/internal/web"
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
			prev, _ := web.SaveConfigPreview(st)
			fmt.Println(prev)
		}
		if err := web.Apply(st, dry); err != nil {
			return err
		}
		fmt.Println("Applied.")
		return nil
	},
}

func init() {
	applyCmd.Flags().Bool("dry-run", false, "preview changes without applying")
}
