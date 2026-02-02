package netutil

import (
	"net"
	"testing"
)

func TestUDPPortAvailable(t *testing.T) {
	pc, err := net.ListenPacket("udp4", "0.0.0.0:0")
	if err != nil {
		t.Fatal(err)
	}
	defer pc.Close()
	addr := pc.LocalAddr().(*net.UDPAddr)
	port := addr.Port
	if UDPPortAvailable(port) {
		t.Fatalf("expected port %d unavailable", port)
	}
}
