package cli

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

var AppDir = homeDir() + "/.gonpm"
var logger = newLogger(AppDir + "/gonpm.log")
var Stdout io.Writer = os.Stdout
var Stderr io.Writer = os.Stderr
var exitFn = os.Exit

func newLogger(path string) *log.Logger {
	err := os.MkdirAll(filepath.Dir(path), 0777)
	must(err)
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	must(err)
	return log.New(file, "", log.LstdFlags)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
