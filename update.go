package main

import (
	"compress/gzip"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/heroku/hk/cli"
)

var hkPath = filepath.Join(cli.AppDir, "hk")

func updateIfNeeded() {
	// TODO: update plugins
	manifest := getUpdateManifest()
	if manifest.Version == Version {
		return
	}
	if !updatable() {
		cli.Errf("Out of date: You are running %s but %s is out.\n", Version, manifest.Version)
		return
	}
	cli.Errf("Updating to %s... ", manifest.Version)
	build := manifest.Builds[runtime.GOOS][runtime.GOARCH]
	tmp, err := downloadHk(build.Url)
	if err != nil {
		panic(err)
	}
	if fileSha1(tmp) != build.Sha1 {
		panic("SHA mismatch")
	}
	if err := os.Rename(tmp, hkPath); err != nil {
		panic(err)
	}
	cli.Errln("done")
	if err := syscall.Exec(hkPath, os.Args, os.Environ()); err != nil {
		panic(err)
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
	path, err := filepath.Abs(os.Args[0])
	if err != nil {
		cli.Errln(err)
	}
	return path == hkPath
}

func downloadHk(url string) (string, error) {
	out, err := os.OpenFile(hkPath+"~", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return "", err
	}
	defer out.Close()
	client := &http.Client{}
	req, err := http.NewRequest("GET", url+".gz", nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept-Encoding", "gzip")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	uncompressed, err := gzip.NewReader(resp.Body)
	if err != nil {
		return "", err
	}
	_, err = io.Copy(out, uncompressed)
	return out.Name(), err
}

func fileSha1(path string) string {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", sha1.Sum(data))
}
