// +build windows

package hkclient

import "os"

const netrcFilename = "_netrc"

func homePath() string {
	home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
	if home == "" {
		home = os.Getenv("USERPROFILE")
	}
	return home
}
