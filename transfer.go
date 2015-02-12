package main

import (
	"io"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/heroku/hk/Godeps/_workspace/src/github.com/bgentry/heroku-go"
)

var cmdTransfer = &Command{
	Run:      runTransfer,
	Usage:    "transfer <email or org name>",
	NeedsApp: true,
	Category: "app",
	Short:    "transfer app ownership to a collaborator or an org" + extra,
	Long: `
Transfer an app's ownership to a collaborator or a Heroku
organization.

Examples:

    $ hk transfer user@test.com
    Requested transfer of myapp to user@test.com.

    $ hk transfer myorg
    Transferred ownership of myapp to myorg.
`,
}

func runTransfer(cmd *Command, args []string) {
	appname := mustApp()
	if len(args) != 1 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	recipient := args[0]

	// if this app has no org AND it's being transferred to another user (email)
	// then we use the regular app transfer endpoint, otherwise use the org
	// endpoint.
	if !isAppInOrg(mustGetOrgApp(appname)) && strings.Contains(recipient, "@") {
		xfer, err := client.AppTransferCreate(appname, recipient)
		must(err)
		log.Printf("Requested transfer of %s to %s.", xfer.App.Name, xfer.Recipient.Email)
	} else {
		_, err := client.OrganizationAppTransferToAccount(appname, recipient)
		must(err)
		log.Printf("Transferred ownership of %s to %s.", appname, recipient)
	}
}

var cmdTransfers = &Command{
	Run:      runTransfers,
	Usage:    "transfers",
	NeedsApp: true,
	Category: "app",
	Short:    "list existing app transfer requests" + extra,
}

func runTransfers(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	transfers, err := client.AppTransferList(nil)
	must(err)

	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()
	for i := range transfers {
		listTransfer(w, transfers[i])
	}
}

func listTransfer(w io.Writer, t heroku.AppTransfer) {
	listRec(w,
		t.App.Name,
		abbrev(t.Owner.Email, 10),
		abbrev(t.Recipient.Email, 10),
		t.State,
		prettyTime{t.UpdatedAt},
	)
}

var cmdTransferAccept = &Command{
	Run:      runTransferAccept,
	Usage:    "transfer-accept",
	NeedsApp: true,
	Category: "app",
	Short:    "accept an inbound app transfer request" + extra,
}

func runTransferAccept(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	xfer, err := client.AppTransferUpdate(mustApp(), "accepted")
	must(err)
	log.Printf("Accepted transfer of %s from %s.", xfer.App.Name, xfer.Recipient.Email)
}

var cmdTransferDecline = &Command{
	Run:      runTransferDecline,
	Usage:    "transfer-decline",
	NeedsApp: true,
	Category: "app",
	Short:    "decline an inbound app transfer request" + extra,
}

func runTransferDecline(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	xfer, err := client.AppTransferUpdate(mustApp(), "declined")
	must(err)
	log.Printf("Declined transfer of %s to %s.", xfer.App.Name, xfer.Recipient.Email)
}

var cmdTransferCancel = &Command{
	Run:      runTransferCancel,
	Usage:    "transfer-cancel",
	NeedsApp: true,
	Category: "app",
	Short:    "cancel an outbound app transfer request" + extra,
}

func runTransferCancel(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	appname := mustApp()
	must(client.AppTransferDelete(appname))
	log.Printf("Canceled transfer of %s.", appname)
}
