package app

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func MustBeRoot() error {
	if os.Geteuid() != 0 {
		return ErrNotRoot
	}
	return nil
}

func RandToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func NowRFC3339() string { return time.Now().UTC().Format(time.RFC3339) }

// Exec runs a command and returns stdout+stderr combined.
// IMPORTANT: Never pass secrets via args; use env or stdin when possible.
func Exec(name string, args ...string) (string, int, error) {
	cmd := exec.Command(name, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	code := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else {
			code = -1
		}
	}
	return buf.String(), code, err
}

func CommandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func EnsureDir(path string, perm os.FileMode) error {
	if err := os.MkdirAll(path, perm); err != nil {
		return err
	}
	return os.Chmod(path, perm)
}

func AtomicWriteFile(path string, perm os.FileMode, data []byte) error {
	dir := filepath.Dir(path)
	tmp := filepath.Join(dir, fmt.Sprintf(".%s.tmp.%d", filepath.Base(path), time.Now().UnixNano()))
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func CopyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := EnsureDir(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func Mask(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 6 {
		return "***"
	}
	return s[:2] + strings.Repeat("*", len(s)-4) + s[len(s)-2:]
}
