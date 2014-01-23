// +build darwin freebsd linux netbsd openbsd

package main

import (
	"os"
	"os/user"
	"path/filepath"
	"syscall"
)

func defaultPluginPath() string {
	hkpath := "/usr/local/lib/hk/plugin"
	fi, err := os.Stat(hkpath)
	if err == nil && fi.IsDir() {
		return hkpath
	}
	if u, err := user.Current(); err != nil {
		return filepath.Join(".", ".hk", "plugins")
	} else {
		return filepath.Join(u.HomeDir, ".hk", "plugins")
	}
}

func sysExec(path string, args []string, env []string) error {
	return syscall.Exec(path, args, env)
}
