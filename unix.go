// +build darwin freebsd linux netbsd openbsd

package main

import (
	"os"
)

var (
	homePath = os.Getenv("HOME")
)
