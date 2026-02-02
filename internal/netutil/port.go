package netutil

import (
	"fmt"
	"net"
	"time"
)

func UDPPortAvailable(port int) bool {
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	pc, err := net.ListenPacket("udp4", addr)
	if err == nil {
		_ = pc.SetDeadline(time.Now().Add(50 * time.Millisecond))
		_ = pc.Close()
		return true
	}
	// try v6 as well (best-effort)
	addr6 := fmt.Sprintf("[::]:%d", port)
	pc6, err6 := net.ListenPacket("udp6", addr6)
	if err6 == nil {
		_ = pc6.SetDeadline(time.Now().Add(50 * time.Millisecond))
		_ = pc6.Close()
		return true
	}
	return false
}

func ChoosePort(preferred int, candidates []int) int {
	if preferred > 0 && UDPPortAvailable(preferred) {
		return preferred
	}
	for _, p := range candidates {
		if UDPPortAvailable(p) {
			return p
		}
	}
	// last resort: ephemeral scan
	for p := 20000; p <= 65000; p++ {
		if UDPPortAvailable(p) {
			return p
		}
	}
	return preferred
}
