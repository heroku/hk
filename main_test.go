package main

import (
	"net/http"
	"os"
	"testing"
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
	os.Setenv("HEROKU_API_URL", "https://api.otherheroku.com")
	initClients()
	os.Setenv("HEROKU_API_URL", "")
}
