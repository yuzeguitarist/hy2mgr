package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/yuzeguitarist/hy2mgr/internal/app"
)

type Settings struct {
	ListenHost        string `json:"listenHost"`        // server public host/IP (for URI), empty -> auto detect
	ListenPort        int    `json:"listenPort"`        // UDP port
	SNI               string `json:"sni"`               // default SNI for clients
	MasqueradeURL     string `json:"masqueradeUrl"`     // reverse proxy target
	MasqueradeRewrite bool   `json:"masqueradeRewrite"` // rewriteHost
	ManageListen      string `json:"manageListen"`      // web UI bind, default 0.0.0.0:3333
	ManagePublic      bool   `json:"managePublic"`      // if true, bind to 0.0.0.0 (explicit)
}

type Admin struct {
	Username     string `json:"username"`
	PasswordBcrypt string `json:"passwordBcrypt"` // bcrypt hash
	TOTPEnabled  bool   `json:"totpEnabled"`
	TOTPSecret   string `json:"totpSecret"` // base32; stored root-only
}

type Node struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username"` // used for hysteria userpass map key
	Password  string `json:"password"` // stored root-only; never log
	Enabled   bool   `json:"enabled"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type Subscription struct {
	// Store only SHA-256 of token. Token itself is shown once on creation/rotation.
	TokenSHA256 string `json:"tokenSha256"`
	CreatedAt   string `json:"createdAt"`
	RevokedAt   string `json:"revokedAt,omitempty"`
}

type State struct {
	Version      int          `json:"version"`
	Settings     Settings     `json:"settings"`
	Admin        Admin        `json:"admin"`
	Nodes        []Node       `json:"nodes"`
	Subscription Subscription `json:"subscription"`
	mu           sync.Mutex   `json:"-"`
}

func Default() *State {
	return &State{
		Version: 1,
		Settings: Settings{
			ListenPort:        443,
			SNI:               "www.bing.com",
			MasqueradeURL:     "https://www.bing.com",
			MasqueradeRewrite: true,
			ManageListen:      "0.0.0.0:3333",
			ManagePublic:      false,
		},
		Admin: Admin{
			Username: "admin",
		},
	}
}

func LoadOrInit() (*State, error) {
	_ = app.EnsureDir(app.StateDir, 0700)
	_ = app.EnsureDir(app.StateBackups, 0700)

	if _, err := os.Stat(app.StatePath); errors.Is(err, os.ErrNotExist) {
		st := Default()
		return st, nil
	}
	b, err := os.ReadFile(app.StatePath)
	if err != nil {
		return nil, err
	}
	var st State
	if err := json.Unmarshal(b, &st); err != nil {
		return nil, err
	}
	if st.Version == 0 {
		st.Version = 1
	}
	return &st, nil
}

func (s *State) SaveAtomic() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	// backup
	if _, err := os.Stat(app.StatePath); err == nil {
		backup := filepath.Join(app.StateBackups, filepath.Base(app.StatePath)+"."+app.NowRFC3339()+".bak")
		_ = os.WriteFile(backup, b, 0600)
	}
	return app.AtomicWriteFile(app.StatePath, 0600, b)
}

func (s *State) NodesSorted() []Node {
	cp := append([]Node{}, s.Nodes...)
	sort.Slice(cp, func(i, j int) bool { return cp[i].CreatedAt < cp[j].CreatedAt })
	return cp
}
