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
		log.Fatal("Invalid usage. See 'hk help domain-add'")
	}
	_, err := client.DomainCreate(mustApp(), args[0])
	must(err)
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
		log.Fatal("Invalid usage. See 'hk help domain-remove'")
	}
	must(client.DomainDelete(mustApp(), args[0]))
}
