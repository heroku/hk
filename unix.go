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
	home := ""
	u, err := user.Current()
	if err != nil {
		home = "."
	} else {
		home = u.HomeDir
	}
	return filepath.Join(home, ".hk", "plugins")
}

func sysExec(path string, args []string, env []string) error {
	return syscall.Exec(path, args, env)
}
