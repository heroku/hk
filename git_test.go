package main

import (
	"os"
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
}

func testAppNameFromGitURL(t *testing.T) {

	os.Setenv("HEROKU_GIT_HOST", "heroku\\.com(\\..*)")

	if res := appNameFromGitURL("git@heroku.com.company_name:myapp.git"); res != "company_name" {
		t.Errorf("expected company_name, got %s", res)
	}

	os.Setenv("HEROKU_GIT_HOST_REGEX", "heroku(\\..*)\\.com")

	if res := appNameFromGitURL("git@heroku.account.com:myapp.git"); res != "account" {
		t.Errorf("expected account, got %s", res)
	}

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
