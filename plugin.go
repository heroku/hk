package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"syscall"
)

func maybeExecPlugin(args []string) {
	var path string
	path = lookupPlugin(args[0])
	if path == "" {
		if path = lookupPlugin("default"); path == "" {
			fmt.Fprintf(os.Stderr, "Unknown command/plugin: %s\n", args[0])
			usage()
		}
	}

	u, err := url.Parse(apiURL)
	if err != nil {
		log.Fatal(err)
	}

	hkuser, hkpass := getCreds(u)
	hkapp, _ := app()
	env := []string{
		"HKAPP=" + hkapp,
		"HKUSER=" + hkuser,
		"HKPASS=" + hkpass,
		"HKHOST=" + u.Host,
		"HKVERSION=" + Version,
	}

	defer os.Exit(2)
	err = syscall.Exec(path, args, append(env, os.Environ()...))
	log.Println("exec error: ", err)
}

// NOTE: lookupPlugin is not threadsafe for anything needing the PATH env var.
func lookupPlugin(name string) string {
	const defaultPluginPath = "/usr/local/lib/hk/plugin"

	opath := os.Getenv("PATH")
	defer os.Setenv("PATH", opath)

	os.Setenv("PATH", os.Getenv("HKPATH")+":"+defaultPluginPath)
	path, err := exec.LookPath(name)
	if err != nil {
		if e, ok := err.(*exec.Error); ok && e.Err == exec.ErrNotFound {
			return ""
		}
		log.Fatal(err)
	}
	return path
}
