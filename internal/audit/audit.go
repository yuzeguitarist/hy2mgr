package audit

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/example/hy2mgr/internal/app"
)

type Entry struct {
	Time   string `json:"time"`
	IP     string `json:"ip"`
	User   string `json:"user"`
	Action string `json:"action"`
	Object string `json:"object,omitempty"`
	Detail string `json:"detail,omitempty"`
}

func Write(e Entry) {
	_ = app.EnsureDir(app.AuditDir, 0750)
	_ = os.Chmod(app.AuditDir, 0750)
	// file 0640 root:adm best-effort
	b, _ := json.Marshal(e)
	path := app.AuditPath
	_ = os.MkdirAll(filepath.Dir(path), 0750)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(append(b, '\n'))
}
