package cli

import (
	"os/user"
	"path/filepath"
)

var HomeDir = homeDir()
var AppDir = filepath.Join(HomeDir, ".hk")

func homeDir() string {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	return user.HomeDir
}
