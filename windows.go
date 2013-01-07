// +build windows

package main

import (
	"os"
)

var (
	homePath = os.Getenv("%HOMEPATH%")
)
