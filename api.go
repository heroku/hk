package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

type Request http.Request

func APIReq(meth, path string) *Request {
	req, err := http.NewRequest(meth, apiURL+path, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.SetBasicAuth(getCreds(req.URL))
	req.Header.Add("User-Agent", "hk/"+Version+" ("+runtime.GOOS+"-"+runtime.GOARCH+")")
	req.Header.Add("Accept", "application/json")
	for _, h := range strings.Split(os.Getenv("HKHEADER"), "\n") {
		i := strings.Index(h, ":")
		if i >= 0 {
			req.Header.Add(h[:i], strings.TrimSpace(h[i+1:]))
		}
	}
	return (*Request)(req)
}

func (r *Request) SetBodyJson(data interface{}) {
	body, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	r.SetBody(bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
}

func (r *Request) SetBodyForm(data url.Values) {
	r.SetBody(strings.NewReader(data.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
}

func (r *Request) SetBody(body io.Reader) {
	rc, ok := body.(io.ReadCloser)
	if !ok && body != nil {
		rc = ioutil.NopCloser(body)
	}
	r.Body = rc
	if body != nil {
		switch v := body.(type) {
		case *strings.Reader:
			r.ContentLength = int64(v.Len())
		case *bytes.Buffer:
			r.ContentLength = int64(v.Len())
		}
	}
}

func (r *Request) Do(v interface{}) {
	res := checkResp(http.DefaultClient.Do((*http.Request)(r)))
	defer res.Body.Close()

	if v != nil {
		err := json.NewDecoder(res.Body).Decode(v)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func checkResp(res *http.Response, err error) *http.Response {
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode == 401 {
		log.Fatal("Unauthorized")
	}
	if res.StatusCode == 403 {
		log.Fatal("Unauthorized")
	}
	if res.StatusCode/100 != 2 { // 200, 201, 202, etc
		log.Fatal("Unexpected error: ", res.Status)
	}

	if msg := res.Header.Get("X-Heroku-Warning"); msg != "" {
		fmt.Fprintln(os.Stderr, strings.TrimSpace(msg))
	}

	return res
}
