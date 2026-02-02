package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yuzeguitarist/hy2mgr/internal/app"
	"github.com/yuzeguitarist/hy2mgr/internal/systemd"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore /etc/hysteria/config.yaml from a backup created by hy2mgr",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := app.MustBeRoot(); err != nil {
			return err
		}
		backup, _ := cmd.Flags().GetString("backup")
		if backup == "" {
			backup = latestBackup()
			if backup == "" {
				return fmt.Errorf("no backup found")
			}
		}
		if !strings.HasPrefix(backup, "/") {
			backup = filepath.Join(filepath.Dir(app.HysteriaConfigPath), backup)
		}
		fmt.Println("Restoring from:", backup)
		if err := app.CopyFile(backup, app.HysteriaConfigPath, 0640); err != nil {
			return err
		}
		_ = systemd.Restart(app.HysteriaService)
		fmt.Println("Restored and restarted.")
		return nil
	},
}

func latestBackup() string {
	dir := filepath.Dir(app.HysteriaConfigPath)
	entries, _ := os.ReadDir(dir)
	var cands []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), filepath.Base(app.HysteriaConfigPath)+".") && strings.Contains(e.Name(), ".bak") {
			cands = append(cands, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(cands)
	if len(cands) == 0 {
		return ""
	}
	return cands[len(cands)-1]
}

func init() {
	restoreCmd.Flags().String("backup", "", "backup file name or absolute path (default: latest)")
}
