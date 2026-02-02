package web

import "github.com/yuzeguitarist/hy2mgr/internal/qr"

func QrPNG(content string, size int) ([]byte, error) { return qr.PNG(content, size) }
func QrSVG(content string, ppm int) ([]byte, error)  { return qr.SVG(content, ppm) }
