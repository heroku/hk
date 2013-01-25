package main

import (
	"bytes"
	"code.google.com/p/go-netrc/netrc"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	repoURL   = "https://github.com/kr/hk.git"
	buildPlat = runtime.GOOS + "-" + runtime.GOARCH
	dir       = "hk"
)

func mustHaveEnv(name string) {
	if os.Getenv(name) == "" {
		log.Fatal("need env: " + name)
	}
}

func build() {
	mustHaveEnv("S3DISTURL")
	mustHaveEnv("S3_ACCESS_KEY")
	mustHaveEnv("S3_SECRET_KEY")
	mustHaveEnv("BUILDBRANCH")
	mustHaveEnv("BUILDNAME")
	mustHaveEnv("DISTURL")
	mustCmd("rm", "-rf", dir)
	mustCmd("git", "clone", "-b", branch, repoURL, dir)
	if err := os.Chdir(dir); err != nil {
		log.Fatal(err)
	}
	ver := mustBuild()
	body, err := os.Open(buildName)
	if err != nil {
		log.Fatal(err)
	}
	sha256 := mustUpload(body, ver)
	mustRegister(ver, sha256)
	mustSetCur(ver)
}

const relverGo = `
// +build release

package main
const Version = %q
`

func mustBuild() (ver string) {
	tag := string(bytes.TrimSpace(mustCmd("git", "describe")))
	if tag[0] != 'v' {
		log.Fatal("bad tag name: ", tag)
	}
	ver = tag[1:]
	if strings.IndexFunc(ver, badVersionRune) >= 0 {
		log.Fatal("bad tag name: ", tag)
	}
	// TODO(kr): verify signature
	url := distURL + buildName + "-" + ver + "-" + buildPlat + ".json"
	if _, err := fetchBytes(url); err == nil {
		log.Fatal("already built: ", ver)
	}

	f, err := os.Create("relver.go")
	if err != nil {
		log.Fatal(err)
	}
	_, err = fmt.Fprintf(f, relverGo, ver)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("go build -tags release -o " + buildName)
	cmd := exec.Command("go", "build", "-tags", "release", "-o", buildName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatal("go build -tags release: ", err)
	}
	return ver
}

func mustUpload(r io.Reader, ver string) (hash []byte) {
	buf := new(bytes.Buffer)
	gz, _ := gzip.NewWriterLevel(buf, gzip.BestCompression)
	gz.Name = buildName + "-" + ver
	if _, err := io.Copy(gz, r); err != nil {
		log.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		log.Fatal(err)
	}
	h := sha256.New()
	h.Write(buf.Bytes())
	hash = h.Sum(nil)
	filename := buildName + "-" + ver + "-" + buildPlat + ".gz"
	if err := s3put(buf, s3DistURL+filename); err != nil {
		log.Fatal(err)
	}
	return hash
}

func mustCmd(arg ...string) []byte {
	output, err := cmd(arg...)
	if err != nil {
		log.Fatal(strings.Join(arg, " ")+": ", err)
	}
	return output
}

func cmd(arg ...string) ([]byte, error) {
	log.Println(strings.Join(arg, " "))
	cmd := exec.Command(arg[0], arg[1:]...)
	cmd.Stderr = os.Stderr
	return cmd.Output()
}

func getCreds(u *url.URL) (user, pass string) {
	if u.User != nil {
		pw, _ := u.User.Password()
		return u.User.Username(), pw
	}

	m, err := netrc.FindMachine(netrcPath, u.Host)
	if err != nil {
		log.Fatalf("netrc error (%s): %v", u.Host, err)
	}

	return m.Login, m.Password
}

func mustRegister(ver string, sha256 []byte) {
	url := distURL + buildName + "-" + ver + "-" + buildPlat + ".json"
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(struct{ Sha256 []byte }{sha256})
	if err != nil {
		log.Fatal(err)
	}
	r, err := http.NewRequest("PUT", url, b)
	if err != nil {
		log.Fatal(err)
	}
	r.SetBasicAuth(getCreds(r.URL))
	r.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Fatal("error in mustRegister:", err)
	}
	if resp.StatusCode != 201 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Fatalf("http status %v putting %q: %q", resp.Status, r.URL, string(body))
	}
}

func mustSetCur(ver string) {
	url := distURL + buildName + "-" + buildPlat + ".json"
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(struct{ Version string }{ver})
	if err != nil {
		log.Fatal(err)
	}
	r, err := http.NewRequest("PUT", url, b)
	if err != nil {
		log.Fatal(err)
	}
	r.SetBasicAuth(getCreds(r.URL))
	r.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Fatalf("http status %v putting %q: %q", resp.Status, r.URL, string(body))
	}
}
