package main

import (
	"io"
	"log"
	"os"
	"text/tabwriter"

	"github.com/bgentry/heroku-go"
)

var cmdTransfer = &Command{
	Run:      runTransfer,
	Usage:    "transfer <email>",
	NeedsApp: true,
	Category: "app",
	Short:    "transfer app ownership to a collaborator" + extra,
}

func runTransfer(cmd *Command, args []string) {
	appname := mustApp()
	if len(args) != 1 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	recipient := args[0]
	xfer, err := client.AppTransferCreate(appname, recipient)
	must(err)
	log.Printf("Requested transfer of %s to %s.", xfer.App.Name, xfer.Recipient.Email)
}

var cmdTransfers = &Command{
	Run:      runTransfers,
	Usage:    "transfers",
	NeedsApp: true,
	Category: "app",
	Short:    "list existing app transfers" + extra,
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
	Short:    "accept an inbound app transfer" + extra,
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
	Short:    "decline an inbound app transfer" + extra,
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
	Short:    "cancel an outbound app transfer" + extra,
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
