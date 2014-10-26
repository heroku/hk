package plugins

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/dickeyxxx/gonpm/cli"
)

var nodePath = filepath.Join(cli.AppDir, NODE_STRING, "bin", "node")
var npmPath = filepath.Join(cli.AppDir, NODE_STRING, "bin", "npm")

func fileExists(path string) (bool, error) {
	var err error
	if runtime.GOOS == "windows" {
		// Windows doesn't seem to like using os.Stat
		_, err = os.Open(path)
	} else {
		_, err = os.Stat(path)
	}
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
