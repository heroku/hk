package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"github.com/kr/binarydist"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"time"
)

var magic = [8]byte{'h', 'k', 'D', 'I', 'F', 'F', '0', '1'}

const upcktimePath = "cktime"

func readTimestamp(path string) time.Time {
	p, err := ioutil.ReadFile(path)
	if err != nil {
		// if the file is missing, treat it as old
		if os.IsNotExist(err) {
			return time.Time{}
		}

		// for any other error, treat it as new
		return nowUTC().Add(1000 * time.Hour)
	}

	t, err := time.Parse(time.RFC3339, string(p))
	if err != nil {
		return nowUTC().Add(1000 * time.Hour)
	}
	return t
}

func writeTimestamp(path string, d time.Duration) bool {
	t := nowUTC().Add(d)
	return ioutil.WriteFile(path, []byte(t.Format(time.RFC3339)), 0644) == nil
}

func nowUTC() time.Time {
	return time.Now().UTC()
}

type Updater struct {
	url string
	dir string
}

func (u *Updater) run() {
	os.MkdirAll(u.dir, 0777)
	if u.wantUpdate() {
		exec.Command("hk", "update").Start()
	}
}

func (u *Updater) wantUpdate() bool {
	wait := 24*time.Hour + time.Duration(rand.Int63n(int64(24*time.Hour)))
	return u.enabled() &&
		nowUTC().After(readTimestamp(u.dir+upcktimePath)) &&
		writeTimestamp(u.dir+upcktimePath, wait)
}

func (u *Updater) enabled() bool {
	_, err := os.Stat(hkHome+"/noupdate")
	return err != nil
}

func (u *Updater) fetchAndApply() error {
	instPath, err := exec.LookPath("hk")
	if err != nil {
		return err
	}

	old, err := os.Open(instPath)
	if err != nil {
		return err
	}

	plat := runtime.GOOS + "-" + runtime.GOARCH
	name := "hk-" + Version + "-next-" + plat + ".hkdiff"
	resp, err := http.Get(u.url + name)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatal(resp.Status)
	}

	var header struct {
		Magic    [8]byte
		OldHash  [sha1.Size]byte
		NewHash  [sha1.Size]byte
		DiffHash [sha1.Size]byte
	}
	err = binary.Read(resp.Body, binary.BigEndian, &header)
	if err != nil {
		return err
	}

	if header.Magic != magic {
		log.Fatal("format error in update file")
	}

	patch, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if readSha1(old) != header.OldHash {
		log.Fatal("existing version hash match update")
	}

	if readSha1(bytes.NewReader(patch)) != header.DiffHash {
		log.Fatal("bad patch file")
	}

	_, err = old.Seek(0, 0)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = binarydist.Patch(old, &buf, bytes.NewReader(patch))
	if err != nil {
		return err
	}

	if readSha1(bytes.NewReader(buf.Bytes())) != header.NewHash {
		log.Fatal("checksum mismatch after patch")
	}

	part := path.Dir(instPath) + "/.hk.part"
	dstf, err := os.OpenFile(part, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	defer os.Remove(part)

	_, err = dstf.Write(buf.Bytes())
	if err != nil {
		return err
	}

	err = dstf.Close()
	if err != nil {
		return err
	}

	return os.Rename(part, instPath)
}

func readSha1(r io.Reader) (h [sha1.Size]byte) {
	s := sha1.New()
	if _, err := io.Copy(s, r); err == nil {
		s.Sum(h[:0])
	}
	return h
}
