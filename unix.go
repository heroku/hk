// +build darwin freebsd linux netbsd openbsd

package main

import (
	"os"
	"path/filepath"
	"syscall"
)

const (
	netrcFilename           = ".netrc"
	acceptPasswordFromStdin = true
)

func sysExec(path string, args []string, env []string) error {
	return syscall.Exec(path, args, env)
}

func homePath() string {
	return os.Getenv("HOME")
}

func defaultPluginPath() string {
	hkpath := "/usr/local/lib/hk/plugin"
	if fi, err := os.Stat(hkpath); err == nil && fi.IsDir() {
		return hkpath
    }
	return filepath.Join(homePath(), ".hk", "plugins")
}
