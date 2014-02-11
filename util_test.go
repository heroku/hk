package main

import (
	"log"
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

	// test with a nil machine
	u, p = getCreds("https://someotherapi.heroku.com")
	if u != "" || p != "" {
		t.Errorf("expected empty user and pass, got u=%q p=%q", u, p)
	}
}

func TestNetrcPath(t *testing.T) {
	fakepath := "/fake/net/rc"
	os.Setenv("NETRC_PATH", fakepath)
	if p := netrcPath(); p != fakepath {
		t.Errorf("NETRC_PATH override expected %q, got %q", fakepath, p)
	}
	os.Setenv("NETRC_PATH", "")
}

func TestStringsIndex(t *testing.T) {
	a1 := []string{}
	if res := stringsIndex(a1, ""); res != -1 {
		t.Errorf("expected -1, got %d", res)
	}
	if res := stringsIndex(a1, "-a"); res != -1 {
		t.Errorf("expected -1, got %d", res)
	}

	a2 := []string{"-a", "bbq"}
	if res := stringsIndex(a2, "-a"); res != 0 {
		t.Errorf("expected 0, got %d", res)
	}
	if res := stringsIndex(a2, "bbq"); res != 1 {
		t.Errorf("expected 1, got %d", res)
	}
	if res := stringsIndex(a2, ""); res != -1 {
		t.Errorf("expected -1, got %d", res)
	}
}
