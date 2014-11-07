package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/heroku/hk/cli"
)

func updateIfNeeded() {
	// TODO: update plugins
	manifest := getUpdateManifest()
	if manifest.Version != Version {
		if !updatable() {
			cli.Errf("Out of date: You are running %s but %s is out.\n", Version, manifest.Version)
			return
		}
		build := manifest.Builds[runtime.GOOS][runtime.GOARCH]
		update(build.Url, build.Sha1)
	}
}

type manifest struct {
	Channel, Version string
	Builds           map[string]map[string]struct {
		Url, Sha1 string
	}
}

func getUpdateManifest() manifest {
	channel := "dev"
	res, err := http.Get("https://d1gvo455cekpjp.cloudfront.net/hk/" + channel + "/manifest.json")
	if err != nil {
		panic(err)
	}
	var m manifest
	json.NewDecoder(res.Body).Decode(&m)
	return m
}

func updatable() bool {
	return true
	return cli.AppDir == runningInDirectory()
}

func runningInDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		cli.Errln(err)
	}
	return dir
}

func update(url, sha1 string) {
	fmt.Println("Updating to", url)
}
