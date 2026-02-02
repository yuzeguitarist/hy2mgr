package hysteria

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Listen     string     `yaml:"listen,omitempty"`
	TLS        TLSConfig  `yaml:"tls"`
	Auth       AuthConfig `yaml:"auth"`
	Masquerade *Masq      `yaml:"masquerade,omitempty"`
}

type TLSConfig struct {
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

type AuthConfig struct {
	Type     string            `yaml:"type"`
	Password string            `yaml:"password,omitempty"`
	Userpass map[string]string `yaml:"userpass,omitempty"`
}

type Masq struct {
	Type       string    `yaml:"type"`
	Proxy      *MasqProxy `yaml:"proxy,omitempty"`
	ListenHTTP  string   `yaml:"listenHTTP,omitempty"`
	ListenHTTPS string   `yaml:"listenHTTPS,omitempty"`
	ForceHTTPS  bool     `yaml:"forceHTTPS,omitempty"`
}

type MasqProxy struct {
	URL         string `yaml:"url"`
	RewriteHost bool   `yaml:"rewriteHost"`
	Insecure    bool   `yaml:"insecure,omitempty"`
}

func GenerateYAML(listenPort int, certPath, keyPath string, users map[string]string, masqueradeURL string, rewriteHost bool) ([]byte, error) {
	cfg := ServerConfig{
		Listen: fmt.Sprintf(":%d", listenPort),
		TLS: TLSConfig{
			Cert: certPath,
			Key:  keyPath,
		},
		Auth: AuthConfig{
			Type:     "userpass",
			Userpass: users,
		},
		Masquerade: &Masq{
			Type: "proxy",
			Proxy: &MasqProxy{
				URL:         masqueradeURL,
				RewriteHost: rewriteHost,
				Insecure:    false,
			},
		},
	}
	// schema per official docs citeturn2view1turn4view0
	return yaml.Marshal(&cfg)
}

func ValidateYAML(y []byte) error {
	var cfg ServerConfig
	if err := yaml.Unmarshal(y, &cfg); err != nil {
		return err
	}
	if cfg.TLS.Cert == "" || cfg.TLS.Key == "" {
		return fmt.Errorf("tls.cert/tls.key required")
	}
	if cfg.Auth.Type == "" {
		return fmt.Errorf("auth.type required")
	}
	if cfg.Auth.Type == "userpass" && len(cfg.Auth.Userpass) == 0 {
		return fmt.Errorf("auth.userpass required for userpass mode")
	}
	return nil
}
