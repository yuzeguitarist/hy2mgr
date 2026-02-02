package qr

import (
	"bytes"
	"fmt"

	qrcode "github.com/skip2/go-qrcode"
)

func PNG(content string, size int) ([]byte, error) {
	return qrcode.Encode(content, qrcode.Medium, size)
}

func SVG(content string, pixelsPerModule int) ([]byte, error) {
	qr, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return nil, err
	}
	bitmap := qr.Bitmap()
	n := len(bitmap)
	if n == 0 {
		return nil, fmt.Errorf("empty qr")
	}
	w := n * pixelsPerModule
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, w, w, w, w))
	buf.WriteString(`<rect width="100%" height="100%" fill="white"/>`)
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			if bitmap[y][x] {
				buf.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="black"/>`, x*pixelsPerModule, y*pixelsPerModule, pixelsPerModule, pixelsPerModule))
			}
		}
	}
	buf.WriteString(`</svg>`)
	return buf.Bytes(), nil
}
