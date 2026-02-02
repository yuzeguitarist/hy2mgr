package cmd

import (
	"fmt"
	"os"

	"github.com/example/hy2mgr/internal/app"
	"github.com/example/hy2mgr/internal/state"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "hy2mgr",
		Short: "HY2 Manager - Hysteria2 server + node manager (single-binary, safe defaults)",
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func mustLoadState() *state.State {
	st, err := state.LoadOrInit()
	if err != nil {
		fmt.Println("failed to load state:", err)
		os.Exit(1)
	}
	_ = app.EnsureDir(app.StateDir, 0700)
	return st
}

func init() {
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(restoreCmd)

	rootCmd.AddCommand(nodeCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(certCmd)
	rootCmd.AddCommand(webCmd)
}
