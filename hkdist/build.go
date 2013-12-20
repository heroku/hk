package main

import (
	"bytes"
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
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bgentry/go-netrc/netrc"
	"github.com/bgentry/heroku-go"
)

const numgen = 20

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

var (
	hkuser        string
	hkpass        string
	identityToken string
	client        heroku.Client
)

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
	mustHaveEnv("S3PATCHURL")
	mustHaveEnv("S3_ACCESS_KEY")
	mustHaveEnv("S3_SECRET_KEY")
	mustHaveEnv("BUILDBRANCH")
	mustHaveEnv("BUILDNAME")
	mustHaveEnv("DISTURL")
	mustHaveEnv("HKGENAPPNAME")

	// determine list of platforms to be built
	platforms := allPlatforms
	if len(args) > 0 {
		platforms = args
	}

	// clone repo
	dir := buildName + "-build"
	err := cloneRepo("https://github.com/heroku/hk.git", buildbranch, dir)
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
	if ver[0] != 'v' || strings.IndexFunc(ver[1:], badVersionRune) >= 0 {
		log.Fatalf("bad tag name: %s", ver)
	}

	// TODO(kr): verify signature

	hkuser, hkpass = getCreds("api.heroku.com")
	client.Username = hkuser
	client.Password = hkpass
	desc := fmt.Sprintf("%s release %s", buildName, ver)
	// get Heroku OAuth token to provide to the hkdist API
	identityToken, err = identityauthreq(desc, []string{"identity"})
	if err != nil {
		log.Fatalf("provision identity oauth authorization: %s", err)
	}

	var dgroup sync.WaitGroup
	dchan := make(chan diff)
	// spawn diff generators
	for i := 0; i < numgen; i++ {
		go func() {
			for d := range dchan {
				d.Generate()
				dgroup.Done()
			}
		}()
	}

	// run Build for each platform
	allSuccessful := true
	builds := make([]*Build, 0)
	for _, platform := range platforms {
		sepIndex := strings.Index(platform, "-")
		b := &Build{
			Name: buildName,
			OS:   platform[:sepIndex],
			Arch: platform[sepIndex+1:],
			Ver:  ver[1:],
		}

		err = b.EnsureBuiltAndRegistered()
		if err != nil {
			allSuccessful = false
			log.Println(err)
			continue
		}
		builds = append(builds, b) // only add to release list if successful

		// Generate diffs
		b.GenDiffs(dchan, &dgroup)
	}
	dgroup.Wait()
	close(dchan)

	for _, b := range builds {
		if err = b.setCurVersion(); err != nil {
			allSuccessful = false
			log.Printf("setCurVersion: %s", err)
		}
	}
	if !allSuccessful {
		os.Exit(1)
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

func (b *Build) EnsureBuiltAndRegistered() error {
	var sha []byte
	// Check if it's already registered
	if registered, err := b.alreadyRegistered(); err != nil {
		return fmt.Errorf("checking registration %s on %s for %s: %s", b.Name, b.Ver, b.platform(), err)
	} else if registered {
		log.Printf("already registered %s on %s for %s", b.Name, b.Ver, b.platform())
	} else {
		sha, err = b.buildAndUpload()
		if err != nil {
			return fmt.Errorf("building %s on %s for %s: %s", b.Name, b.Ver, b.platform(), err)
		}
		if err = b.register(sha); err != nil {
			return fmt.Errorf("registration: %s", err)
		}
	}
	return nil
}

func (b *Build) buildAndUpload() (shasum []byte, err error) {
	if err = b.build(); err != nil {
		return
	}
	body, err := os.Open(b.filename())
	if err != nil {
		return
	}

	h := sha256.New()
	if _, err = io.Copy(h, body); err != nil {
		return
	}
	shasum = h.Sum(nil)

	if _, err = body.Seek(int64(0), 0); err != nil {
		return
	}

	err = b.upload(body)
	return
}

const relverGo = `
// +build release

package main
const (
	Version = %q
)

var updater = &Updater{
	apiURL:  %q,
	binURL:  %q,
	diffURL: %q,
	dir:     hkHome() + "/update/",
	cmdName: %q,
}
`

func (b *Build) build() (err error) {
	log.Printf("building cmd=%s release=%s os=%s arch=%s\n", b.Name, b.Ver, b.OS, b.Arch)
	f, err := os.Create("relver.go")
	if err != nil {
		return fmt.Errorf("writing relver.go: %s", err)
	}
	_, err = fmt.Fprintf(f, relverGo, b.Ver, distURL, s3DistURL, s3PatchURL, b.Name)
	if err != nil {
		return fmt.Errorf("writing relver.go: %s", err)
	}
	cmd := exec.Command("godep", "go", "build", "-tags", "release", "-o", b.filename())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	env := []string{"GOOS=" + b.OS, "GOARCH=" + b.Arch, "CGO_ENABLED=0"}
	cmd.Env = append(env, os.Environ()...)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("godep go build -tags release: ", err)
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

	if err := s3put(buf, b.url()); err != nil {
		return err
	}
	return nil
}

func (b *Build) url() string {
	return s3DistURL + b.Name + "/" + b.Ver + "/" + b.platform() + ".gz"
}

func (b *Build) alreadyRegistered() (bool, error) {
	url := distURL + b.Name + "/" + b.Ver + "/" + b.platform() + ".json"
	if resp, err := http.Head(url); err != nil {
		return false, err
	} else {
		return resp.StatusCode == 200, nil
	}
}

func (b *Build) register(sha256 []byte) error {
	url := distURL + b.Name + "/" + b.Ver + "/" + b.platform() + ".json"
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(jsonsha{sha256})
	if err != nil {
		return err
	}
	r, err := http.NewRequest("PUT", url, buf)
	if err != nil {
		return err
	}
	r.SetBasicAuth(hkuser, identityToken)
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
	url := distURL + b.Name + "/current/" + b.platform() + ".json"
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(struct{ Version string }{b.Ver})
	if err != nil {
		return err
	}
	r, err := http.NewRequest("PUT", url, buf)
	if err != nil {
		return err
	}
	r.SetBasicAuth(hkuser, identityToken)
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

func (b *Build) GenDiffs(dchan chan diff, dgroup *sync.WaitGroup) {
	diffs, err := b.getDiffs()
	if err != nil {
		// TODO: Log error
		log.Printf("Build.getDiffs release=%s os=%s arch=%s msg=%q\n", b.Ver, b.OS, b.Arch, err)
		return
	}

	// Add diff gens to work queue
	dgroup.Add(len(diffs))
	for _, d := range diffs {
		dchan <- d
	}
}

func (b *Build) getDiffs() ([]diff, error) {
	versions, err := b.getOldVersions(8)
	if err != nil {
		return nil, err
	}
	diffs := make([]diff, len(versions))
	for i, ver := range versions {
		diffs[i] = diff{Cmd: b.Name, Platform: b.platform(), From: ver, To: b.Ver}
	}
	return diffs, nil
}

func (b *Build) getOldVersions(limit int) ([]string, error) {
	url := distURL + "release.json"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error fetching releases: %d", resp.StatusCode)
	}
	var rels []release
	if err = json.NewDecoder(resp.Body).Decode(&rels); err != nil {
		return nil, err
	}
	var versions sort.StringSlice
	for _, r := range rels {
		if r.Cmd == b.Name && r.Plat == b.platform() && r.Ver != b.Ver {
			versions = append(versions, r.Ver)
		}
	}

	sort.Sort(sort.Reverse(versions))
	if len(versions) < limit {
		limit = len(versions)
	}
	return versions[:limit], nil
}

type diff struct {
	Cmd      string
	Platform string
	From     string
	To       string
}

func (d *diff) Exists() bool {
	// Check if diff already exists
	urlstr := s3PatchURL + patchFilename(d.Cmd, d.Platform, d.From, d.To)
	if resp, err := http.Head(urlstr); err != nil {
		ue, ok := err.(*url.Error)
		if !ok || ue.Err != io.EOF {
			log.Printf("diff.Exists name=%s platform=%s from=%s to=%s error=%q", d.Cmd, d.Platform, d.From, d.To, ue.Err)
		}
		return false
	} else {
		return resp.StatusCode == 200
	}
}

func (d *diff) Generate() {
	if d.Exists() {
		return
	}

	d.runGen(time.Now().Add(45 * time.Second))
}

func (d *diff) runGen(deadline time.Time) {
	command := "hkdist gen " + d.Cmd + " " + d.Platform + " " + d.From + " " + d.To
	_, err := client.DynoCreate(hkgenAppName, command, nil)
	if err != nil {
		log.Printf("diff.runGen %s -> %s: %s", d.From, d.To, err)
		return
	}

	for _ = range time.Tick(5 * time.Second) {
		if time.Now().After(deadline) {
			log.Printf("diff.runGen timeout cmd=%s platform=%s from=%s to=%s", d.Cmd, d.Platform, d.From, d.To)
			return
		}

		if d.Exists() {
			return
		}
	}
}

func cmd(arg ...string) ([]byte, error) {
	log.Println(strings.Join(arg, " "))
	cmd := exec.Command(arg[0], arg[1:]...)
	cmd.Stderr = os.Stderr
	return cmd.Output()
}

func getCreds(host string) (user, pass string) {
	m, err := netrc.FindMachine(netrcPath, host)
	if err != nil {
		log.Fatalf("netrc error (%s): %v", host, err)
	}

	return m.Login, m.Password
}

func runreq(appname, command string) error {
	_, err := client.DynoCreate(appname, command, nil)
	return err
}

func identityauthreq(desc string, scopes []string) (string, error) {
	expires := 600
	opts := heroku.OAuthAuthorizationCreateOpts{
		Description: &desc,
		ExpiresIn:   &expires,
	}
	auth, err := client.OAuthAuthorizationCreate(scopes, &opts)
	if err != nil {
		return "", err
	}

	return auth.AccessToken.Token, nil
}
