// +build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

const netrcFilename = "_netrc"

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

func homePath() string {
	home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
	if home == "" {
		home = os.Getenv("USERPROFILE")
	}
	return home
}

func defaultPluginPath() string {
	return filepath.Join(homePath(), ".hk", "plugins")
}
