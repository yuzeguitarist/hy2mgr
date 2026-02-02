package cmd

import (
	"fmt"
	"strings"

	"github.com/yuzeguitarist/hy2mgr/internal/app"
	"github.com/yuzeguitarist/hy2mgr/internal/service"
	"github.com/spf13/cobra"
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Manage nodes (accounts)",
}

var nodeAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a node",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := app.MustBeRoot(); err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("--name required")
		}
		st := mustLoadState()
		n, err := service.NodeAdd(st, name, "", "")
		if err != nil {
			return err
		}
		fmt.Println("Node created:", n.ID, n.Name)
		fmt.Println("URI: hy2mgr export uri --id", n.ID)
		fmt.Println("QR : hy2mgr export qrcode --id", n.ID, "--out ./"+n.ID+".png")
		return nil
	},
}

var nodeRmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove a node",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := app.MustBeRoot(); err != nil {
			return err
		}
		id, _ := cmd.Flags().GetString("id")
		if id == "" {
			return fmt.Errorf("--id required")
		}
		st := mustLoadState()
		if err := service.NodeDelete(st, id); err != nil {
			return err
		}
		fmt.Println("Deleted:", id)
		return nil
	},
}

var nodeLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List nodes",
	RunE: func(cmd *cobra.Command, args []string) error {
		st := mustLoadState()
		fmt.Printf("%-10s  %-18s  %-8s  %s\n", "ID", "USERNAME", "ENABLED", "NAME")
		for _, n := range st.NodesSorted() {
			fmt.Printf("%-10s  %-18s  %-8v  %s\n", n.ID, n.Username, n.Enabled, n.Name)
		}
		return nil
	},
}

var nodeDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable a node",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := app.MustBeRoot(); err != nil {
			return err
		}
		id, _ := cmd.Flags().GetString("id")
		st := mustLoadState()
		if err := service.NodeSetEnabled(st, id, false); err != nil {
			return err
		}
		fmt.Println("Disabled:", id)
		return nil
	},
}

var nodeEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable a node",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := app.MustBeRoot(); err != nil {
			return err
		}
		id, _ := cmd.Flags().GetString("id")
		st := mustLoadState()
		if err := service.NodeSetEnabled(st, id, true); err != nil {
			return err
		}
		fmt.Println("Enabled:", id)
		return nil
	},
}

var nodeResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset node password (new password generated)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := app.MustBeRoot(); err != nil {
			return err
		}
		id, _ := cmd.Flags().GetString("id")
		st := mustLoadState()
		if err := service.NodeResetPassword(st, id); err != nil {
			return err
		}
		fmt.Println("Reset password:", id)
		fmt.Println("Next: hy2mgr export uri --id", id)
		return nil
	},
}

func init() {
	nodeCmd.AddCommand(nodeAddCmd, nodeRmCmd, nodeLsCmd, nodeDisableCmd, nodeEnableCmd, nodeResetCmd)
	nodeAddCmd.Flags().String("name", "", "node display name")
	nodeRmCmd.Flags().String("id", "", "node id")
	nodeDisableCmd.Flags().String("id", "", "node id")
	nodeEnableCmd.Flags().String("id", "", "node id")
	nodeResetCmd.Flags().String("id", "", "node id")
}
