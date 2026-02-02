package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuzeguitarist/hy2mgr/internal/app"
	"github.com/yuzeguitarist/hy2mgr/internal/qr"
	"github.com/yuzeguitarist/hy2mgr/internal/service"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export client artifacts (URI, QR code, subscription)",
}

var exportURICmd = &cobra.Command{
	Use:   "uri",
	Short: "Print hysteria2 URI for a node",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("id")
		if id == "" {
			return fmt.Errorf("--id required")
		}
		st := mustLoadState()
		uri, err := service.NodeURI(st, id)
		if err != nil {
			return err
		}
		fmt.Println(uri)
		return nil
	},
}

var exportQRCmd = &cobra.Command{
	Use:   "qrcode",
	Short: "Export QR code for a node URI (PNG or SVG by file extension)",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("id")
		out, _ := cmd.Flags().GetString("out")
		if id == "" || out == "" {
			return fmt.Errorf("--id and --out required")
		}
		st := mustLoadState()
		uri, err := service.NodeURI(st, id)
		if err != nil {
			return err
		}
		var b []byte
		if strings.HasSuffix(strings.ToLower(out), ".svg") {
			b, err = qr.SVG(uri, 6)
		} else {
			b, err = qr.PNG(uri, 256)
		}
		if err != nil {
			return err
		}
		if err := os.WriteFile(out, b, 0644); err != nil {
			return err
		}
		fmt.Println("Wrote:", filepath.Clean(out))
		return nil
	},
}

var exportSubCmd = &cobra.Command{
	Use:   "subscription",
	Short: "Print subscription URL and optionally rotate token",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := app.MustBeRoot(); err != nil {
			return err
		}
		rotate, _ := cmd.Flags().GetBool("rotate")
		st := mustLoadState()
		if rotate || st.Subscription.TokenSHA256 == "" {
			token, path, err := service.SubscriptionRotate(st)
			if err != nil {
				return err
			}
			fmt.Println("New token (shown once):", token)
			fmt.Println("Subscription URL:", "http://127.0.0.1:3333"+path)
			return nil
		}
		fmt.Println("Token is stored hashed; to show a usable URL, rotate it:")
		fmt.Println("  hy2mgr export subscription --rotate")
		return nil
	},
}

func init() {
	exportCmd.AddCommand(exportURICmd, exportQRCmd, exportSubCmd)
	exportURICmd.Flags().String("id", "", "node id")
	exportQRCmd.Flags().String("id", "", "node id")
	exportQRCmd.Flags().String("out", "", "output file path (.png or .svg)")
	exportSubCmd.Flags().Bool("rotate", false, "rotate token and print the new URL")
}
