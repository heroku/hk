package main

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/bgentry/heroku-go"
)

var cmdDomains = &Command{
	Run:      runDomains,
	Usage:    "domains",
	NeedsApp: true,
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
	Usage:    "domain-add <domain>",
	NeedsApp: true,
	Category: "domain",
	Short:    "add a domain",
}

func runDomainAdd(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	appname := mustApp()
	domain := args[0]
	_, err := client.DomainCreate(appname, domain)
	must(err)
	log.Printf("Added %s to %s.", domain, appname)
}

var cmdDomainRemove = &Command{
	Run:      runDomainRemove,
	Usage:    "domain-remove <domain>",
	NeedsApp: true,
	Category: "domain",
	Short:    "remove a domain",
}

func runDomainRemove(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	appname := mustApp()
	domain := args[0]
	must(client.DomainDelete(appname, domain))
	log.Printf("Removed %s from %s.", domain, appname)
}
