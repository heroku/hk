package main

import (
	"fmt"
	"github.com/bgentry/heroku-go"
	"log"
	"os"
	"text/tabwriter"
)

var cmdDomains = &Command{
	Run:      runDomains,
	Name:     "domains",
	Usage:    "[-a <app>]",
	Category: "domain",
	Short:    "list domains",
	Long: `
Lists domains.

Examples:

    $ hk domains
    test.herokuapp.com
    www.test.com
`,
}

func init() {
	cmdDomains.Flag.StringVar(&flagApp, "a", "", "app name")
}

func runDomains(cmd *Command, args []string) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	appname := mustApp()
	domains, err := client.DomainList(appname, &heroku.ListRange{
		Field: "hostname",
		Max:   1000,
	})
	must(err)

	for _, d := range domains {
		fmt.Fprintln(w, d.Hostname)
	}
}

var cmdDomainAdd = &Command{
	Run:      runDomainAdd,
	Name:     "domain-add",
	Usage:    "[-a <app>] <domain>",
	Category: "domain",
	Short:    "add a domain",
}

func init() {
	cmdDomainAdd.Flag.StringVar(&flagApp, "a", "", "app name")
}

func runDomainAdd(cmd *Command, args []string) {
	if len(args) != 1 {
		log.Fatal("Invalid usage. See 'hk help domain-add'")
	}
	_, err := client.DomainCreate(mustApp(), args[0])
	must(err)
}

var cmdDomainRemove = &Command{
	Run:      runDomainRemove,
	Name:     "domain-remove",
	Usage:    "[-a <app>] <domain>",
	Category: "domain",
	Short:    "remove a domain",
}

func init() {
	cmdDomainRemove.Flag.StringVar(&flagApp, "a", "", "app name")
}

func runDomainRemove(cmd *Command, args []string) {
	if len(args) != 1 {
		log.Fatal("Invalid usage. See 'hk help domain-remove'")
	}
	must(client.DomainDelete(mustApp(), args[0]))
}
