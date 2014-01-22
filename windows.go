// +build windows

package main

import (
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
)

func defaultPluginPath() string {
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
