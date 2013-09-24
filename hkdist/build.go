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
	"strings"
	"time"
)

var allPlatforms = []string{
	"darwin-386",
	"darwin-amd64",
	"freebsd-386",
	"freebsd-amd64",
	"freebsd-arm",
	"linux-386",
	"linux-amd64",
	"linux-arm",
	"windows-386",
	"windows-amd64",
}

func mustHaveEnv(name string) {
	if os.Getenv(name) == "" {
		log.Fatal("need env: " + name)
	}
}

func cloneRepo(repo, branch, dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	if _, err := cmd("git", "clone", "-b", branch, repo, dir); err != nil {
		return err
	}
	return nil
}

func build(args []string) {
	mustHaveEnv("S3DISTURL")
	mustHaveEnv("S3_ACCESS_KEY")
	mustHaveEnv("S3_SECRET_KEY")
	mustHaveEnv("BUILDBRANCH")
	mustHaveEnv("BUILDNAME")
	mustHaveEnv("DISTURL")

	// determine list of platforms to be built
	platforms := allPlatforms
	if len(args) > 0 {
		platforms = args
	}

	// clone repo
	dir := buildName + "-build"
	err := cloneRepo("https://github.com/kr/hk.git", buildbranch, dir)
	if err != nil {
		log.Fatalf("cloning repo to %s on branch %s: %s\n", dir, buildbranch, err)
	}

	if err := os.Chdir(dir); err != nil {
		log.Fatal(err)
	}

	tagb, err := cmd("git", "describe")
	if err != nil {
		log.Fatalf("listing tags: %s", err)
	}
	ver := string(bytes.TrimSpace(tagb))
	if (ver[0] != 'v' || strings.IndexFunc(ver[1:], badVersionRune) >= 0) {
		log.Fatalf("bad tag name: %s", ver)
	}

	// TODO(kr): verify signature

	// run Build for each platform
	for _, platform := range platforms {
		sepIndex := strings.Index(platform, "-")
		b := &Build{
			Name: buildName,
			OS:   platform[:sepIndex],
			Arch: platform[sepIndex+1:],
			Ver:  ver[1:],
		}
		err := b.Run()
		if err != nil {
			log.Printf("building %s on %s for %s: %s\n", b.Name, b.Ver, b.platform(), err)
		}
	}
}

type Build struct {
	Name string
	OS   string
	Arch string
	Ver  string
}

func (b *Build) filename() string {
	if b.OS == "windows" {
		return b.Name + ".exe"
	}
	return b.Name
}

func (b *Build) platform() string {
	return b.OS + "-" + b.Arch
}

func (b *Build) Run() error {
	url := distURL + b.Name + "-" + b.Ver + "-" + b.platform() + ".json"
	if resp, err := http.Head(url); err == nil && resp.StatusCode == 200 {
		return fmt.Errorf("already built: %s", b.Ver)
	}

	err := b.build()
	if err != nil {
		return err
	}
	body, err := os.Open(b.filename())
	if err != nil {
		return err
	}

	h := sha256.New()
	if _, err := io.Copy(h, body); err != nil {
		return err
	}
	shasum := h.Sum(nil)

	_, err = body.Seek(int64(0), 0)
	if err != nil {
		return err
	}

	if err = b.upload(body); err != nil {
		return fmt.Errorf("upload: %s", err)
	}
	if err = b.register(shasum); err != nil {
		return fmt.Errorf("registration: %s", err)
	}
	if err = b.setCurVersion(); err != nil {
		return fmt.Errorf("release: %s", err)
	}
	return nil
}

const relverGo = `
// +build release

package main
const Version = %q
`

func (b *Build) build() (err error) {
	log.Printf("building release=%s os=%s arch=%s\n", b.Ver, b.OS, b.Arch)
	f, err := os.Create("relver.go")
	if err != nil {
		return fmt.Errorf("writing relver.go: %s", err)
	}
	_, err = fmt.Fprintf(f, relverGo, b.Ver)
	if err != nil {
		return fmt.Errorf("writing relver.go: %s", err)
	}
	log.Printf("GOOS=%s GOARCH=%s go build -tags release -o %s\n", b.OS, b.Arch, b.filename())
	cmd := exec.Command("go", "build", "-tags", "release", "-o", b.filename())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	env := []string{"GOOS="+b.OS, "GOARCH="+b.Arch, "CGO_ENABLED=0"}
	cmd.Env = append(env, os.Environ()...)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("go build -tags release: ", err)
	}
	return nil
}

func (b *Build) upload(r io.Reader) error {
	buf := new(bytes.Buffer)
	gz, _ := gzip.NewWriterLevel(buf, gzip.BestCompression)
	gz.Name = b.Name + "-" + b.Ver
	if b.OS == "windows" {
		gz.Name += ".exe"
	}
	if _, err := io.Copy(gz, r); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}

	filename := b.Name + "/" + b.Ver + "/" + b.platform() + ".gz"
	if err := s3put(buf, s3DistURL+filename); err != nil {
		return err
	}
	return nil
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

func (b *Build) register(sha256 []byte) error {
	url := distURL + b.Name + "-" + b.Ver + "-" + b.platform() + ".json"
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(struct{ Sha256 []byte }{sha256})
	if err != nil {
		return err
	}
	r, err := http.NewRequest("PUT", url, buf)
	if err != nil {
		return err
	}
	r.SetBasicAuth(getCreds(r.URL))
	r.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	if resp.StatusCode != 201 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("http status %v putting %q: %q", resp.Status, r.URL, string(body))
	}
	return nil
}

func (b *Build) setCurVersion() error {
	url := distURL + b.Name + "-" + b.platform() + ".json"
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(struct{ Version string }{b.Ver})
	if err != nil {
		return err
	}
	r, err := http.NewRequest("PUT", url, buf)
	if err != nil {
		return err
	}
	r.SetBasicAuth(getCreds(r.URL))
	r.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("http status %v putting %q: %q", resp.Status, r.URL, string(body))
	}
	return nil
}
