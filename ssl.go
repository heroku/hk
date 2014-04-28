package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/bgentry/heroku-go"
)

var cmdSSL = &Command{
	Run:      runSSL,
	Usage:    "ssl",
	NeedsApp: true,
	Category: "ssl",
	Short:    "show ssl endpoint info",
	Long:     `Show SSL endpoint and certificate information.`,
}

func runSSL(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	endpoints, err := client.SSLEndpointList(mustApp(), nil)
	must(err)

	if len(endpoints) == 0 {
		return
	}

	chain, err := decodeCertChain(endpoints[0].CertificateChain)
	must(err)

	fmt.Println("Hostname:       ", endpoints[0].Cname)
	fmt.Println("Common Name(s): ", strings.Join(chain.CommonNames(), ", "))
	fmt.Println("Expires:        ", chain.Expires().UTC().Format(time.RFC3339))
}

var cmdSSLCertAdd = &Command{
	Run:      runSSLCertAdd,
	Usage:    "ssl-cert-add [-s] <certfile> <keyfile>",
	NeedsApp: true,
	Category: "ssl",
	Short:    "add a new ssl cert",
	Long: `
Add a new SSL certificate to an app. An SSL endpoint will be
created if the app doesn't yet have one. Otherwise, its cert will
be updated.

Options:

    -s  skip SSL cert optimization and pre-processing

Examples:

    $ hk ssl-cert-add cert.pem key.pem
    hobby-dev        $0/mo
`,
}

var (
	skipCertPreprocess bool
)

func init() {
	cmdSSLCertAdd.Flag.BoolVarP(&skipCertPreprocess, "skip-preprocess", "s", false, "skip SSL cert preprocessing")
}

func runSSLCertAdd(cmd *Command, args []string) {
	if len(args) != 2 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	appname := mustApp()

	certb, err := ioutil.ReadFile(args[0])
	if err != nil {
		printFatal("reading certfile: %s", err.Error())
	}
	keyb, err := ioutil.ReadFile(args[1])
	if err != nil {
		printFatal("reading keyfile: %s", err.Error())
	}

	endpoints, err := client.SSLEndpointList(appname, nil)
	must(err)

	cert := string(certb)
	key := string(keyb)

	preprocess := !skipCertPreprocess

	if len(endpoints) == 0 {
		opts := heroku.SSLEndpointCreateOpts{Preprocess: &preprocess}
		ep, err := client.SSLEndpointCreate(appname, cert, key, &opts)
		must(err)
		fmt.Printf("Added cert for %s at %s.\n", appname, ep.Cname)
		return
	}

	opts := heroku.SSLEndpointUpdateOpts{
		CertificateChain: &cert,
		Preprocess:       &preprocess,
		PrivateKey:       &key,
	}
	_, err = client.SSLEndpointUpdate(appname, endpoints[0].Id, &opts)
	must(err)
	fmt.Printf("Updated cert for %s at %s.\n", appname, endpoints[0].Cname)
}

var cmdSSLDestroy = &Command{
	Run:      runSSLDestroy,
	Usage:    "ssl-destroy",
	NeedsApp: true,
	Category: "ssl",
	Short:    "destroy ssl endpoint",
	Long: `
Removes the SSL endpoints from an app along with all SSL
certificates. If your app's DNS is still configured to point at
the SSL endpoint, this may take your app offline. The command
will prompt for confirmation, or accept confirmation via stdin.

Examples:

    $ hk ssl-destroy
    warning: This will destroy the SSL endpoint on myapp. Please type "myapp" to continue:
    > myapp
    Destroyed SSL endpoint on myapp.

    $ echo myapp | hk ssl-destroy
    Destroyed SSL endpoint on myapp.
`,
}

func runSSLDestroy(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	appname := mustApp()

	endpoints, err := client.SSLEndpointList(appname, nil)
	must(err)

	if len(endpoints) == 0 {
		printFatal("App %s has no SSL endpoint to destroy.", appname)
	}

	warning := "This will destroy the SSL endpoint on %s. Please type %q to continue:"
	mustConfirm(fmt.Sprintf(warning, appname, appname), appname)

	err = client.SSLEndpointDelete(appname, endpoints[0].Id)
	must(err)
	fmt.Printf("Destroyed SSL endpoint on %s.\n", appname)
}

var cmdSSLCertRollback = &Command{
	Run:      runSSLCertRollback,
	Usage:    "ssl-cert-rollback",
	NeedsApp: true,
	Category: "ssl",
	Short:    "add a new ssl cert",
	Long: `
Rolls back an SSL endpoint's certificate to the previous version.

Examples:

    $ hk ssl-cert-rollback
    Rolled back cert for myapp.
`,
}

func runSSLCertRollback(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	appname := mustApp()

	endpoints, err := client.SSLEndpointList(appname, nil)
	must(err)

	if len(endpoints) == 0 {
		printFatal("App %s has no SSL endpoint to rollback.", appname)
	}

	t := true
	opts := heroku.SSLEndpointUpdateOpts{Rollback: &t}
	_, err = client.SSLEndpointUpdate(appname, endpoints[0].Id, &opts)
	must(err)
	fmt.Printf("Rolled back cert for %s.\n", appname)
}
