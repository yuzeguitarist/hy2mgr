package netutil

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// PublicIP best-effort returns a public IP for user-facing output.
func PublicIP() string {
	if ip := publicIPFromInterfaces(); ip != "" {
		return ip
	}
	if ip := publicIPFromHTTP(); ip != "" {
		return ip
	}
	return "YOUR_VPS_IP"
}

func publicIPFromInterfaces() string {
	ifaces, _ := net.Interfaces()
	var v6 string
	for _, iface := range ifaces {
		if (iface.Flags&net.FlagUp) == 0 || (iface.Flags&net.FlagLoopback) != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			ip := extractIP(addr)
			if !isPublicIP(ip) {
				continue
			}
			if ip4 := ip.To4(); ip4 != nil {
				return ip4.String()
			}
			if v6 == "" {
				v6 = ip.String()
			}
		}
	}
	return v6
}

func publicIPFromHTTP() string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.ipify.org", nil)
	if err != nil {
		return ""
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return ""
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 64))
	if err != nil {
		return ""
	}
	ip := strings.TrimSpace(string(b))
	if net.ParseIP(ip) == nil {
		return ""
	}
	return ip
}

func extractIP(addr net.Addr) net.IP {
	switch v := addr.(type) {
	case *net.IPNet:
		return v.IP
	case *net.IPAddr:
		return v.IP
	default:
		s := addr.String()
		if i := strings.IndexByte(s, '/'); i >= 0 {
			s = s[:i]
		}
		return net.ParseIP(s)
	}
}

func isPublicIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() || ip.IsMulticast() || ip.IsUnspecified() {
		return false
	}
	if ip.IsPrivate() {
		return false
	}
	return ip.IsGlobalUnicast()
}
