package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
)

func init() {
	if os.Getenv("HEROKU_SSL_VERIFY") == "disable" {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
}

type Accepter interface {
	Accept() string
}

func Get(v interface{}, path string) error {
	return APIReq(v, "GET", path, nil)
}

func Post(v interface{}, path string, body interface{}) error {
	return APIReq(v, "POST", path, body)
}

func Put(v interface{}, path string, body interface{}) error {
	return APIReq(v, "PUT", path, body)
}

// Sends a Heroku API request and decodes the response into v.
// The type of v determines how to handle the response body:
//
//   nil        body is discarded
//   io.Writer  body is copied directly into v
//   else       body is decoded into v as json
//
// If v implements Accepter, v.Accept() will be used as the HTTP
// Accept header.
//
// The type of body determines how to encode the request:
//
//   nil         no body
//   io.Reader   body is sent verbatim
//   url.Values  body is encoded as application/x-www-form-urlencoded
//   else        body is encoded as application/json
func APIReq(v interface{}, meth, path string, body interface{}) error {
	var err error
	var ctype string
	var rbody io.Reader

	switch t := body.(type) {
	case nil:
	case url.Values:
		rbody = strings.NewReader(t.Encode())
		ctype = "application/x-www-form-urlencoded"
	case io.Reader:
		rbody = t
	default:
		j, err := json.Marshal(body)
		if err != nil {
			log.Fatal(err)
		}
		rbody = bytes.NewReader(j)
		ctype = "application/json"
	}
	req, err := http.NewRequest(meth, apiURL+path, rbody)
	if err != nil {
		return err
	}
	req.SetBasicAuth(getCreds(req.URL))
	if a, ok := v.(Accepter); ok {
		req.Header.Add("Accept", a.Accept())
	} else {
		req.Header.Add("Accept", "application/vnd.heroku+json; version=3")
	}
	req.Header.Add("User-Agent", userAgent())
	req.Header.Set("Content-Type", ctype)
	for _, h := range strings.Split(os.Getenv("HKHEADER"), "\n") {
		i := strings.Index(h, ":")
		if i >= 0 {
			req.Header.Add(h[:i], strings.TrimSpace(h[i+1:]))
		}
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if err = checkResp(res); err != nil {
		return err
	}
	switch t := v.(type) {
	case nil:
	case io.Writer:
		_, err = io.Copy(t, res.Body)
	default:
		err = json.NewDecoder(res.Body).Decode(v)
	}
	return err
}

func checkResp(res *http.Response) error {
	if res.StatusCode == 401 {
		return errors.New("Unauthorized")
	}
	if res.StatusCode == 403 {
		return errors.New("Unauthorized")
	}
	if res.StatusCode/100 != 2 { // 200, 201, 202, etc
		return errors.New("Unexpected error: " + res.Status)
	}
	if msg := res.Header.Get("X-Heroku-Warning"); msg != "" {
		fmt.Fprintln(os.Stderr, strings.TrimSpace(msg))
	}
	return nil
}

func userAgent() string {
	return "hk " + Version + " (" + runtime.GOOS + "-" + runtime.GOARCH + ")"
}
