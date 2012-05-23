package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"github.com/kr/binarydist"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"time"
)

var magic = [8]byte{'h', 'k', 'D', 'I', 'F', 'F', '0', '1'}

const (
	upcktimePath  = "cktime"
	upasktimePath = "asktime"
	upnextPath    = "hk.next"
)

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

// isTerminal returns true if f is a terminal.
func isTerminal(f *os.File) bool {
	cmd := exec.Command("test", "-t", "0")
	cmd.Stdin = f
	return cmd.Run() == nil
}

type Updater struct {
	url string
	dir string
}

func (u *Updater) run() {
	os.MkdirAll(u.dir, 0777)
	switch {
	case u.wantInstall():
		u.askAndInstall()
	case u.wantDownload():
		u.bgFetch()
	}
}

func (u *Updater) wantInstall() bool {
	s, err := os.Stat(u.dir + upnextPath)
	if err != nil {
		return false // no update has been downloaded, or some other error
	}

	if s.Mode()&os.ModeType != 0 { // not a regular file?
		return false
	}

	return nowUTC().After(readTimestamp(u.dir + upasktimePath))
}

func (u *Updater) askAndInstall() {
	// Only try to ask if both stdin and stdout are ttys.
	if !isTerminal(os.Stdin) || !isTerminal(os.Stdout) {
		return
	}

	if !writeTimestamp(u.dir+upasktimePath, time.Hour) {
		// If we can't update the timestamp, we won't be able to
		// rate-limit our user prompt, so don't ask at all.
		return
	}

	instPath, err := exec.LookPath("hk")
	if err != nil {
		return
	}

	ver, err := exec.Command(u.dir+upnextPath, "version").Output()
	if err != nil {
		return
	}
	fmt.Printf("\n\nUpdate hk %s has been downloaded.\n", string(bytes.TrimSpace(ver)))
	fmt.Print("Install? (y/[n]) ")
	line, isPrefix, err := stdin.ReadLine()
	if err != nil || isPrefix {
		return
	}

	if bytes.HasPrefix(bytes.TrimSpace(line), []byte{'y'}) {
		srcf, err := os.Open(u.dir + upnextPath)
		if err != nil {
			error(err.Error())
		}

		instDir := path.Dir(instPath)
		dstf, err := os.OpenFile(instDir+"/.hk.part", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
		if err != nil {
			error(err.Error())
		}

		_, err = io.Copy(dstf, srcf)
		if err != nil {
			error(err.Error())
		}

		srcf.Close()
		err = dstf.Close()
		if err != nil {
			error(err.Error())
		}

		err = os.Rename(instDir+"/.hk.part", instPath)
		if err != nil {
			error(err.Error())
		}

		err = os.Remove(u.dir + upnextPath)
		if err != nil {
			error(err.Error())
		}
	}
}

func (u *Updater) wantDownload() bool {
	if nowUTC().After(readTimestamp(u.dir + upcktimePath)) {
		wait := 24*time.Hour + time.Duration(rand.Int63n(int64(24*time.Hour)))
		if !writeTimestamp(u.dir+upcktimePath, wait) {
			return false
		}

		_, err := os.Stat(u.dir + upnextPath)
		return err != nil
	}
	return false
}

func (u *Updater) bgFetch() {
	exec.Command("hk", "fetch-update").Start()
}

func (u *Updater) fetchAndApply() {
	instPath, err := exec.LookPath("hk")
	if err != nil {
		error(err.Error())
	}

	old, err := os.Open(instPath)
	if err != nil {
		error(err.Error())
	}

	plat := runtime.GOOS + "-" + runtime.GOARCH
	resp, err := http.Get(u.url + plat + "-" + Version + "-next.hkdiff")
	if err != nil {
		error(err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		error(resp.Status)
	}

	var header struct {
		Magic    [8]byte
		OldHash  [sha1.Size]byte
		NewHash  [sha1.Size]byte
		DiffHash [sha1.Size]byte
	}
	err = binary.Read(resp.Body, binary.BigEndian, &header)
	if err != nil {
		error(err.Error())
	}

	if header.Magic != magic {
		error("format error in update file")
	}

	patch, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		error(err.Error())
	}

	if !sha1matches(header.OldHash, old) {
		error("existing version hash match update")
	}

	if !sha1matches(header.DiffHash, bytes.NewReader(patch)) {
		error("bad patch file")
	}

	_, err = old.Seek(0, 0)
	if err != nil {
		error(err.Error())
	}

	part := u.dir + upnextPath + ".part"
	newPart, err := os.OpenFile(part, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		error(err.Error())
	}

	err = binarydist.Patch(old, newPart, bytes.NewReader(patch))
	if err != nil {
		error(err.Error())
	}

	err = newPart.Close()
	if err != nil {
		error(err.Error())
	}

	newPart, err = os.Open(part)
	if err != nil {
		error(err.Error())
	}
	if !sha1matches(header.NewHash, newPart) {
		error("checksum mismatch after patch")
	}

	err = os.Rename(part, u.dir+upnextPath)
	if err != nil {
		error(err.Error())
	}
}

func sha1matches(h [sha1.Size]byte, r io.Reader) bool {
	var n [sha1.Size]byte
	s := sha1.New()
	_, err := io.Copy(s, r)
	if err != nil {
		return false
	}
	s.Sum(n[:0])
	return h == n
}
