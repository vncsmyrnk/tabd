package internal

import "path/filepath"

func DataDir() string {
	return filepath.Join("/", "usr", "local", "share", "seman")
}

func SystemdDataDir() string {
	return filepath.Join(DataDir(), "systemd")
}
