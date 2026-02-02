package web

import (
	"strings"

	"github.com/pquerna/otp/totp"
)

func verifyTOTP(secret, code string) bool {
	code = strings.TrimSpace(code)
	if code == "" {
		return false
	}
	return totp.Validate(code, secret)
}
