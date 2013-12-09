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
	Usage:    "domains [-l]",
	Category: "domain",
	Short:    "list domains",
	Long: `
Lists domains.

Options:

    -l       long listing

Long listing for shows the name, state, age, and command.

Examples:

    $ hk domains
    test.herokuapp.com
    www.test.com

    $ hk domains -l
    test.herokuapp.com  Jun 12 18:28  01234567-89ab-cdef-0123-456789abcdef
    www.test.com        Jun 13 18:14  abcdef01-89ab-cdef-9876-543210fedcba
`,
}

func init() {
	cmdDomains.Flag.BoolVar(&flagLong, "l", false, "long listing")
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
		if flagLong {
			listRec(w,
				d.Hostname,
				prettyTime{d.CreatedAt},
				d.Id,
			)
		} else {
			fmt.Fprintln(w, d.Hostname)
		}
	}
}

var cmdDomainAdd = &Command{
	Run:      runDomainAdd,
	Usage:    "domain-add <domain>",
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
	Category: "domain",
	Short:    "remove a domain",
}

func runDomainRemove(cmd *Command, args []string) {
	if len(args) != 1 {
		log.Fatal("Invalid usage. See 'hk help domain-remove'")
	}
	must(client.DomainDelete(mustApp(), args[0]))
}
