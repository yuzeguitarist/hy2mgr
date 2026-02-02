package app

import "errors"

var (
	ErrNotRoot = errors.New("this command must be run as root")
)
