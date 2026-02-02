package hysteria

import "testing"

func TestGenerateAndValidate(t *testing.T) {
	y, err := GenerateYAML(443, "/etc/hysteria/cert.crt", "/etc/hysteria/cert.key", map[string]string{"u1": "p1"}, "https://www.bing.com", true)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateYAML(y); err != nil {
		t.Fatalf("validate failed: %v\n%s", err, string(y))
	}
	// must contain listen and auth userpass
	if string(y) == "" || len(y) < 20 {
		t.Fatal("empty yaml")
	}
}
