package main

import (
	"net/http"
	"os"
	"testing"
)

func TestSSLEnabled(t *testing.T) {
	testSSLEnabledCommand := &Command{
		Run: func(cmd *Command, args []string) {
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
		},
		Usage: "test-ssl-enabled",
	}
	commands = append(commands, testSSLEnabledCommand)
	os.Args = []string{"hk", "test-ssl-enabled"}

	main()
}

func TestSSLDisable(t *testing.T) {
	os.Setenv("HEROKU_SSL_VERIFY", "disable")
	testSSLDisableCommand := &Command{
		Run: func(cmd *Command, args []string) {
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
		},
		Usage: "test-ssl-disabled",
	}
	commands = append(commands, testSSLDisableCommand)
	os.Args = []string{"hk", "test-ssl-disabled"}

	main()

	os.Setenv("HEROKU_SSL_VERIFY", "")
}
