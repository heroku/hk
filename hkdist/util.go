package main

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/heroku/hk/Godeps/_workspace/src/github.com/kr/s3"
)

type errReader struct {
	error
}

func (e errReader) Read(p []byte) (int, error) {
	return 0, e.error
}

func (e errReader) Close() error {
	return e.error
}

type gzReader struct {
	z, r io.ReadCloser
}

func newGzReader(r io.ReadCloser) io.ReadCloser {
	var err error
	g := new(gzReader)
	g.r = r
	g.z, err = gzip.NewReader(r)
	if err != nil {
		return errReader{err}
	}
	return g
}

func (g *gzReader) Read(p []byte) (int, error) {
	return g.z.Read(p)
}

func (g *gzReader) Close() error {
	g.z.Close()
	return g.r.Close()
}

func fetch(url string, mod *time.Time) io.ReadCloser {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return errReader{err}
	}
	if mod != nil {
		req.Header.Add("If-Modified-Since", mod.Format(http.TimeFormat))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errReader{err}
	}
	if mod != nil && resp.StatusCode == 304 {
		return nil
	}
	if resp.StatusCode != 200 {
		err := fmt.Errorf("bad http status from %s: %v", url, resp.Status)
		return errReader{err}
	}
	if s := resp.Header.Get("Last-Modified"); mod != nil && s != "" {
		t, err := time.Parse(http.TimeFormat, s)
		if err == nil {
			*mod = t
		}
	}
	return resp.Body
}

func fetchJSON(url string, mod *time.Time, v interface{}) error {
	r := fetch(url, mod)
	if r == nil {
		return nil
	}
	defer r.Close()
	return json.NewDecoder(r).Decode(v)
}

func s3put(bb *bytes.Buffer, url string) error {
	r, _ := http.NewRequest("PUT", url, bb)
	r.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	r.Header.Set("Content-Disposition", "attachment")
	r.Header.Set("X-Amz-Acl", "public-read")
	r.Header.Set("Content-Md5", b64md5(bb.Bytes()))
	s3.Sign(r, s3keys)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("s3 status %v putting %q: %q", resp.Status, url, string(body))
	}
	return nil
}

// returns the base64-encoded md5 of p
func b64md5(p []byte) string {
	h := md5.New()
	h.Write(p)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

type jsonsha struct {
	Sha256 []byte `json:"sha256"`
}
