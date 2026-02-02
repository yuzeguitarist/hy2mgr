package systemd

import (
	"fmt"
	"strings"

	"github.com/yuzeguitarist/hy2mgr/internal/app"
)

func Systemctl(args ...string) (string, error) {
	out, _, err := app.Exec("systemctl", args...)
	if err != nil {
		return out, err
	}
	return out, nil
}

func IsActive(unit string) (bool, error) {
	out, _ := Systemctl("is-active", unit)
	return strings.TrimSpace(out) == "active", nil
}

func Restart(unit string) error {
	_, err := Systemctl("restart", unit)
	return err
}

func Start(unit string) error {
	_, err := Systemctl("start", unit)
	return err
}

func Stop(unit string) error {
	_, err := Systemctl("stop", unit)
	return err
}

func EnableNow(unit string) error {
	_, err := Systemctl("enable", "--now", unit)
	return err
}

func Status(unit string) (string, error) {
	out, err := Systemctl("status", "--no-pager", unit)
	return out, err
}

func JournalTail(unit string, lines int) (string, error) {
	out, _, err := app.Exec("journalctl", "--no-pager", "-n", fmt.Sprintf("%d", lines), "-u", unit)
	return out, err
}

func Cat(unit string) (string, error) {
	out, err := Systemctl("cat", unit)
	return out, err
}
