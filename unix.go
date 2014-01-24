// +build darwin freebsd linux netbsd openbsd

package main

import (
	"os"
	"syscall"
)

const netrcFilename = ".netrc"

func sysExec(path string, args []string, env []string) error {
	return syscall.Exec(path, args, env)
}

func homePath() string {
	return os.Getenv("HOME")
}
