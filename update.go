package main

import (
	"bitbucket.org/kardianos/osext"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/inconshreveable/go-update"
	"github.com/kr/binarydist"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
)

var cmdUpdate = &Command{
	Run:   runUpdate,
	Usage: "update",
	Long: `
Update downloads and installs the next version of hk.

This command is unlisted, since users never have to run it directly.
`,
}

func runUpdate(cmd *Command, args []string) {
	if updater == nil {
		log.Fatal("Dev builds don't support auto-updates")
	}
	if err := updater.update(); err != nil {
		log.Fatal(err)
	}
}

const (
	upcktimePath = "cktime"
	plat         = runtime.GOOS + "-" + runtime.GOARCH
)

const devValidTime = 7 * 24 * time.Hour

// Update protocol.
//
//   GET hk.heroku.com/hk/current/linux-amd64.json
//
//   200 ok
//   {
//       "Version": "2",
//       "Sha256": "..." // base64
//   }
//
// then
//
//   GET hkpatch.s3.amazonaws.com/hk/1/2/linux-amd64
//
//   200 ok
//   [bsdiff data]
//
// or
//
//   GET hkdist.s3.amazonaws.com/hk/2/linux-amd64.gz
//
//   200 ok
//   [gzipped executable data]
type Updater struct {
	apiURL  string
	cmdName string
	binURL  string
	diffURL string
	dir     string
	info    struct {
		Version string
		Sha256  []byte
	}
}

func (u *Updater) backgroundRun() {
	os.MkdirAll(u.dir, 0777)
	if u.wantUpdate() {
		if err := update.SanityCheck(); err != nil {
			// fail
			return
		}
		self, err := osext.Executable()
		if err != nil {
			// fail update, couldn't figure out path to self
			return
		}
		// TODO(bgentry): logger isn't on Windows. Replace w/ proper error reports.
		l := exec.Command("logger", "-thk")
		c := exec.Command(self, "update")
		if w, err := l.StdinPipe(); err == nil && l.Start() == nil {
			c.Stdout = w
			c.Stderr = w
		}
		c.Start()
	}
}

func (u *Updater) wantUpdate() bool {
	path := u.dir + upcktimePath
	if Version == "dev" || readTime(path).After(time.Now()) {
		return false
	}
	wait := 24*time.Hour + randDuration(24*time.Hour)
	return writeTime(path, time.Now().Add(wait))
}

func (u *Updater) update() error {
	path, err := osext.Executable()
	if err != nil {
		return err
	}
	old, err := os.Open(path)
	if err != nil {
		return err
	}
	defer old.Close()

	err = u.fetchInfo()
	if err != nil {
		return err
	}
	if u.info.Version == Version {
		return nil
	}
	bin, err := u.fetchAndApplyPatch(old)
	if err != nil {
		bin, err = u.fetchBin()
		if err != nil {
			return err
		}
	}

	h := sha256.New()
	h.Write(bin)
	if !bytes.Equal(h.Sum(nil), u.info.Sha256) {
		return errors.New("new file hash mismatch after patch")
	}

	// close the old binary before installing because on windows
	// it can't be renamed if a handle to the file is still open
	old.Close()

	err, errRecover := update.FromStream(bytes.NewBuffer(bin))
	if errRecover != nil {
		return fmt.Errorf("update and recovery errors: %q %q", err, errRecover)
	}
	if err != nil {
		return err
	}
	return nil
}

func (u *Updater) fetchInfo() error {
	r, err := fetch(u.apiURL + u.cmdName + "/current/" + plat + ".json")
	if err != nil {
		return err
	}
	defer r.Close()
	err = json.NewDecoder(r).Decode(&u.info)
	if err != nil {
		return err
	}
	if len(u.info.Sha256) != sha256.Size {
		return errors.New("bad cmd hash in info")
	}
	return nil
}

func (u *Updater) fetchAndApplyPatch(old io.Reader) ([]byte, error) {
	r, err := fetch(u.diffURL + u.cmdName + "/" + Version + "/" + u.info.Version + "/" + plat)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var buf bytes.Buffer
	err = binarydist.Patch(old, &buf, r)
	return buf.Bytes(), err
}

func (u *Updater) fetchBin() ([]byte, error) {
	r, err := fetch(u.binURL + u.cmdName + "/" + u.info.Version + "/" + plat + ".gz")
	if err != nil {
		return nil, err
	}
	defer r.Close()
	buf := new(bytes.Buffer)
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(buf, gz); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// returns a random duration in [0,n).
func randDuration(n time.Duration) time.Duration {
	return time.Duration(rand.Int63n(int64(n)))
}

func fetch(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad http status from %s: %v", url, resp.Status)
	}
	return resp.Body, nil
}

func readTime(path string) time.Time {
	p, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		return time.Time{}
	}
	if err != nil {
		return time.Now().Add(1000 * time.Hour)
	}
	t, err := time.Parse(time.RFC3339, string(p))
	if err != nil {
		return time.Now().Add(1000 * time.Hour)
	}
	return t
}

func writeTime(path string, t time.Time) bool {
	return ioutil.WriteFile(path, []byte(t.Format(time.RFC3339)), 0644) == nil
}
