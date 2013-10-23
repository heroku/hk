// +build darwin freebsd linux netbsd openbsd

package main

import (
	"os"
	"syscall"
)

var (
	homePath = os.Getenv("HOME")
)

func sysExec(path string, args []string, env []string) error {
	return syscall.Exec(path, args, env)
}
