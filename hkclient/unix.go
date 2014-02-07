// +build darwin freebsd linux netbsd openbsd

package hkclient

import "os"

const netrcFilename = ".netrc"

func homePath() string {
	return os.Getenv("HOME")
}
