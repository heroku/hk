// +build windows

package main

import (
	"os"
	"os/exec"
)

var (
	homePath = os.Getenv("%HOMEPATH%")
)

func sysExec(path string, args []string, env []string) error {
	cmd := exec.Command(path, args[1:]...)
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
