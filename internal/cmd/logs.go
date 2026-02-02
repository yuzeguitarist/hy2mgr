package cmd

import (
	"fmt"

	"github.com/example/hy2mgr/internal/app"
	"github.com/example/hy2mgr/internal/systemd"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show recent journal logs for hysteria-server.service",
	RunE: func(cmd *cobra.Command, args []string) error {
		lines, _ := cmd.Flags().GetInt("lines")
		if lines <= 0 || lines > 5000 {
			lines = 200
		}
		out, _ := systemd.JournalTail(app.HysteriaService, lines)
		fmt.Println(out)
		return nil
	},
}

func init() {
	logsCmd.Flags().Int("lines", 200, "number of lines")
}
