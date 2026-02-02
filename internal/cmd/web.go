package cmd

import (
	"fmt"
	"net/http"
	"time"

	"github.com/yuzeguitarist/hy2mgr/internal/app"
	"github.com/yuzeguitarist/hy2mgr/internal/state"
	"github.com/yuzeguitarist/hy2mgr/internal/web"
	"github.com/spf13/cobra"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Run embedded Web UI/API server (foreground)",
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := state.LoadOrInit()
		if err != nil {
			return err
		}
		listen, _ := cmd.Flags().GetString("listen")
		if listen == "" {
			listen = st.Settings.ManageListen
		}
		// session key derived from state path (not secret but stable) + random in state would be better for prod
		sk := []byte("change-me-" + app.StatePath)
		srv := web.NewServer(st, sk)

		httpSrv := &http.Server{
			Addr:              listen,
			Handler:           srv.Router(),
			ReadHeaderTimeout: 5 * time.Second,
		}

		fmt.Println("Listening:", listen)
		return httpSrv.ListenAndServe()
	},
}

func init() {
	webCmd.Flags().String("listen", "", "listen address (default from state: 127.0.0.1:3333)")
}
