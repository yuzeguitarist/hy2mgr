package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yuzeguitarist/hy2mgr/internal/app"
)

// GenerateSelfSigned creates a self-signed ECDSA cert with ONLY IP SANs.
// Rationale: avoid DNS SAN to keep Hysteria's sniGuard default (dns-san) from enforcing SNI matching. citeturn1view0
func GenerateSelfSigned(ipAddrs []net.IP, validDays int) (certPEM, keyPEM []byte, pin string, err error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, "", err
	}

	serialLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serial, err := rand.Int(rand.Reader, serialLimit)
	if err != nil {
		return nil, nil, "", err
	}

	notBefore := time.Now().Add(-5 * time.Minute)
	notAfter := time.Now().Add(time.Duration(validDays) * 24 * time.Hour)

	tpl := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   "hy2-selfsigned",
			Organization: []string{"hy2mgr"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,

		IPAddresses: ipAddrs,
	}

	der, err := x509.CreateCertificate(rand.Reader, &tpl, &tpl, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, "", err
	}

	// pinSHA256 is shown in docs as openssl x509 -fingerprint -sha256 output (colon-separated hex). citeturn6view0
	sum := sha256.Sum256(der)
	pin = strings.ToUpper(hex.EncodeToString(sum[:]))
	// format AA:BB:..
	var parts []string
	for i := 0; i < len(pin); i += 2 {
		parts = append(parts, pin[i:i+2])
	}
	pin = strings.Join(parts, ":")

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})

	keyDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, "", err
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})
	return certPEM, keyPEM, pin, nil
}

func WriteCertFiles(certPEM, keyPEM []byte, ownerUID, ownerGID int) error {
	if err := app.EnsureDir(filepath.Dir(app.HysteriaCertPath), 0750); err != nil {
		return err
	}
	if err := os.WriteFile(app.HysteriaCertPath, certPEM, 0644); err != nil {
		return err
	}
	if err := os.WriteFile(app.HysteriaKeyPath, keyPEM, 0640); err != nil {
		return err
	}
	_ = os.Chown(app.HysteriaCertPath, ownerUID, ownerGID)
	_ = os.Chown(app.HysteriaKeyPath, ownerUID, ownerGID)
	return nil
}

func ParseCertPin(certPath string) (string, error) {
	b, err := os.ReadFile(certPath)
	if err != nil {
		return "", err
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return "", fmt.Errorf("invalid PEM: %s", certPath)
	}
	sum := sha256.Sum256(block.Bytes)
	pin := strings.ToUpper(hex.EncodeToString(sum[:]))
	var parts []string
	for i := 0; i < len(pin); i += 2 {
		parts = append(parts, pin[i:i+2])
	}
	return strings.Join(parts, ":"), nil
}
