// +build darwin freebsd linux netbsd openbsd

package main

import (
	"syscall"
)

func sysExec(path string, args []string, env []string) error {
	return syscall.Exec(path, args, env)
}
