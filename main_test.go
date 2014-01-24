package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func init() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	err = os.Setenv("NETRC_PATH", filepath.Join(wd, "fakenetrc"))
	if err != nil {
		log.Fatal(err)
	}
}

func TestSSLEnabled(t *testing.T) {
	initClients()

	if client.HTTP == nil {
		// No http.Client means the client defaults to SSL enabled
		return
	}
	if client.HTTP.Transport == nil {
		// No transport means the client defaults to SSL enabled
		return
	}
	conf := client.HTTP.Transport.(*http.Transport).TLSClientConfig
	if conf == nil {
		// No TLSClientConfig means the client defaults to SSL enabled
		return
	}
	if conf.InsecureSkipVerify {
		t.Errorf("expected InsecureSkipVerify == false")
	}

	if pgclient.HTTP == nil {
		// No http.Client means the pgclient defaults to SSL enabled
		return
	}
	if pgclient.HTTP.Transport == nil {
		// No transport means the pgclient defaults to SSL enabled
		return
	}
	conf = pgclient.HTTP.Transport.(*http.Transport).TLSClientConfig
	if conf == nil {
		// No TLSClientConfig means the pgclient defaults to SSL enabled
		return
	}
	if conf.InsecureSkipVerify {
		t.Errorf("expected InsecureSkipVerify == false")
	}

	client = nil
	pgclient = nil
}

func TestSSLDisable(t *testing.T) {
	os.Setenv("HEROKU_SSL_VERIFY", "disable")
	initClients()

	if client.HTTP == nil {
		t.Fatalf("client.HTTP not set, expected http.Client")
	}
	if client.HTTP.Transport == nil {
		t.Fatalf("client.HTTP.Transport not set")
	}
	conf := client.HTTP.Transport.(*http.Transport).TLSClientConfig
	if conf == nil {
		t.Fatalf("client.HTTP.Transport's TLSClientConfig is nil")
	}
	if !conf.InsecureSkipVerify {
		t.Errorf("expected InsecureSkipVerify == true")
	}

	if pgclient.HTTP == nil {
		t.Fatalf("pgclient.HTTP not set, expected http.Client")
	}
	if pgclient.HTTP.Transport == nil {
		t.Fatalf("pgclient.HTTP.Transport not set")
	}
	conf = pgclient.HTTP.Transport.(*http.Transport).TLSClientConfig
	if conf == nil {
		t.Fatalf("pgclient.HTTP.Transport's TLSClientConfig is nil")
	}
	if !conf.InsecureSkipVerify {
		t.Errorf("expected InsecureSkipVerify == true")
	}

	os.Setenv("HEROKU_SSL_VERIFY", "")
	client = nil
	pgclient = nil
}

func TestGetCreds(t *testing.T) {
	u, p := getCreds("https://omg:wtf@api.heroku.com")
	if u != "omg" {
		t.Errorf("expected user=omg, got %s", u)
	}
	if p != "wtf" {
		t.Errorf("expected password=wtf, got %s", p)
	}
	u, p = getCreds("https://api.heroku.com")
	if u != "user@test.com" {
		t.Errorf("expected user=user@test.com, got %s", u)
	}
	if p != "faketestpassword" {
		t.Errorf("expected password=faketestpassword, got %s", p)
	}
}

func TestHerokuAPIURL(t *testing.T) {
	os.Setenv("HEROKU_API_URL", "https://api.otherheroku.com")
	initClients()
	os.Setenv("HEROKU_API_URL", "")
}

func TestNetrcPath(t *testing.T) {
	fakepath := "/fake/net/rc"
	os.Setenv("NETRC_PATH", fakepath)
	if p := netrcPath(); p != fakepath {
		t.Errorf("NETRC_PATH override expected %q, got %q", fakepath, p)
	}
	os.Setenv("NETRC_PATH", "")
}
