package main

import (
	"net/http"
	"os"
	"testing"

	"github.com/heroku/hk/postgresql"
)

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

func TestHerokuAPIURL(t *testing.T) {
	newURL := "https://api.otherheroku.com"
	os.Setenv("HEROKU_API_URL", newURL)
	initClients()

	if client.URL != newURL {
		t.Errorf("expected client.URL to be %q, got %q", newURL, client.URL)
	}

	// cleanup
	os.Setenv("HEROKU_API_URL", "")
}

func TestHerokuPostgresqlHost(t *testing.T) {
	newHost := "maciek"
	newURL := "https://" + newHost + ".herokuapp.com" + postgresql.DefaultAPIPath
	os.Setenv("HEROKU_POSTGRESQL_HOST", newHost)
	initClients()

	if pgclient.URL != newURL {
		t.Errorf("expected client.URL to be %q, got %q", newURL, pgclient.URL)
	}
	if pgclient.StarterURL != newURL {
		t.Errorf("expected client.StarterURL to be %q, got %q", newURL, pgclient.StarterURL)
	}

	// cleanup
	os.Setenv("HEROKU_POSTGRESQL_HOST", "")
}

func TestShogun(t *testing.T) {
	newShogun := "will"
	newURL := "https://shogun-" + newShogun + ".herokuapp.com" + postgresql.DefaultAPIPath
	os.Setenv("SHOGUN", newShogun)
	initClients()

	if pgclient.URL != newURL {
		t.Errorf("expected client.URL to be %q, got %q", newURL, pgclient.URL)
	}
	// starter URL should be unchanged
	if pgclient.StarterURL != "" {
		t.Errorf("expected client.StarterURL to be empty, got %q", pgclient.StarterURL)
	}

	// cleanup
	os.Setenv("SHOGUN", "")
}

func TestShogunAndHerokuPostgresqlHost(t *testing.T) {
	newShogun := "fdr"
	newShogunURL := "https://shogun-" + newShogun + ".herokuapp.com" + postgresql.DefaultAPIPath
	newHost := "maciek"
	newHostURL := "https://" + newHost + ".herokuapp.com" + postgresql.DefaultAPIPath
	os.Setenv("HEROKU_POSTGRESQL_HOST", newHost)
	os.Setenv("SHOGUN", newShogun)
	initClients()

	if pgclient.URL != newShogunURL {
		t.Errorf("expected client.URL to be %q, got %q", newShogunURL, pgclient.URL)
	}
	// starter URL should be unchanged
	if pgclient.StarterURL != newHostURL {
		t.Errorf("expected client.StarterURL to be %q, got %q", newHostURL, pgclient.StarterURL)
	}

	// TODO
	// cleanup
	os.Setenv("HEROKU_POSTGRESQL_HOST", "")
	os.Setenv("SHOGUN", "")
}
