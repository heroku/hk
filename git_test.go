package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestGitHost(t *testing.T) {
	if res := gitHost(); res != "heroku.com" {
		t.Errorf("expected heroku.com, got %s", res)
	}

	os.Setenv("HEROKU_GIT_HOST", "notheroku.com")

	if res := gitHost(); res != "notheroku.com" {
		t.Errorf("expected notheroku.com, got %s", res)
	}

	os.Setenv("HEROKU_GIT_HOST", "")
	os.Setenv("HEROKU_HOST", "stillnotheroku.com")
	defer os.Setenv("HEROKU_HOST", "")

	if res := gitHost(); res != "stillnotheroku.com" {
		t.Errorf("expected stillnotheroku.com, got %s", res)
	}

	fi, err := ioutil.TempFile("", "config")
	if err != nil {
		t.Errorf("expected alsonotheroku.com, but couldn't create temp config file")
	}
	if _, err := fi.WriteString("[heroku]\n\thost = alsonotheroku.com\n"); err != nil {
		t.Errorf("expected alsonotheroku.com, but couldn't write temp config file")
	}
	b, _ := exec.Command("git", "config", "--file", fi.Name(), "heroku.host").Output()
	if res := strings.TrimSpace(string(b)); res != "alsonotheroku.com" {
		t.Errorf("expected alsonotheroku.com, got %s", res)
	}
	defer func() {
		fi.Close()
		os.Remove(fi.Name())
	}()

}

var gitRemoteTestOutput = `
heroku	git@heroku.com:myappfetch.git (fetch)
heroku	git@heroku.com:myapp.git (push)
staging	git@heroku.com:myapp-staging.git (fetch)
staging	git@heroku.com:myapp-staging.git (push)
origin	git@github.com:heroku/hk.git (fetch)
origin	git@github.com:heroku/hk.git (push)
`

func TestParseGitRemoteOutput(t *testing.T) {
	results, err := parseGitRemoteOutput([]byte(gitRemoteTestOutput))
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]string{
		"heroku":  "myapp",
		"staging": "myapp-staging",
	}

	if len(results) != len(expected) {
		t.Errorf("expected %d results, got %d", len(expected), len(results))
	}

	for remoteName, app := range expected {
		val, ok := results[remoteName]
		if !ok {
			t.Errorf("expected remote %s not found", val)
		} else if val != app {
			t.Errorf("expected remote %s to point to app %s, got %s", remoteName, app, val)
		}
	}
}
