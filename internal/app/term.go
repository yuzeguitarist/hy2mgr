package app

import "os"

// Color wraps text with ANSI color code when stdout is a terminal and NO_COLOR is not set.
func Color(text, code string) string {
	if code == "" || !colorEnabled() {
		return text
	}
	return "\x1b[" + code + "m" + text + "\x1b[0m"
}

func colorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}
