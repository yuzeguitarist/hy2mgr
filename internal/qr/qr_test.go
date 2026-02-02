package qr

import (
	"bytes"
	"testing"
)

func TestPNG(t *testing.T) {
	b, err := PNG("hello", 128)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) < 8 || !bytes.HasPrefix(b, []byte{0x89, 'P', 'N', 'G'}) {
		t.Fatalf("not png: %v", b[:8])
	}
}

func TestSVG(t *testing.T) {
	b, err := SVG("hello", 4)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(b, []byte("<svg")) {
		t.Fatal("not svg")
	}
}
