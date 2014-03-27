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
		cmd.printUsage()
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
	Usage:    "ssl-cert-add <certfile> <keyfile>",
	NeedsApp: true,
	Category: "ssl",
	Short:    "add a new ssl cert",
	Long: `
Add a new SSL certificate to an app. An SSL endpoint will be
created if the app doesn't yet have one. Otherwise, its cert will
be updated.

Examples:

    $ hk ssl-cert-add cert.pem key.pem
    hobby-dev        $0/mo
`,
}

func runSSLCertAdd(cmd *Command, args []string) {
	if len(args) != 2 {
		cmd.printUsage()
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

	if len(endpoints) == 0 {
		ep, err := client.SSLEndpointCreate(appname, cert, key)
		must(err)
		fmt.Printf("Added cert for %s at %s.\n", appname, ep.Cname)
		return
	}

	opts := heroku.SSLEndpointUpdateOpts{
		CertificateChain: &cert,
		PrivateKey:       &key,
	}
	_, err = client.SSLEndpointUpdate(appname, endpoints[0].Id, &opts)
	must(err)
	fmt.Printf("Updated cert for %s at %s.\n", appname, endpoints[0].Cname)
}
