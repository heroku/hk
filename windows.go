// +build windows

package main

import (
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
)

func defaultPluginPath() string {
	if u, err := user.Current(); err != nil {
		return filepath.Join(".", ".hk", "plugins")
	} else {
		return filepath.Join(u.HomeDir, ".hk", "plugins")
	}
}

func sysExec(path string, args []string, env []string) error {
	cmd := exec.Command(path, args...)
	cmd.Env = env
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err != nil {
		return err
	}
	os.Exit(0)
	return nil
}
