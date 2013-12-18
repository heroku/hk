package main

import (
	"github.com/bgentry/heroku-go"
	"io"
	"log"
	"os"
	"text/tabwriter"
)

var cmdTransfer = &Command{
	Run:      runTransfer,
	Usage:    "transfer <email>",
	NeedsApp: true,
	Category: "app",
	Short:    "transfer app ownership to a collaborator" + extra,
}

func runTransfer(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	recipient := args[0]
	xfer, err := client.AppTransferCreate(mustApp(), recipient)
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
	xfer := mustLookupTransfer(mustApp())
	must(updateTransferState(xfer.Id, "accepted"))
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
	xfer := mustLookupTransfer(mustApp())
	must(updateTransferState(xfer.Id, "declined"))
}

var cmdTransferCancel = &Command{
	Run:      runTransferCancel,
	Usage:    "transfer-cancel",
	NeedsApp: true,
	Category: "app",
	Short:    "cancel an outbound app transfer" + extra,
}

func runTransferCancel(cmd *Command, args []string) {
	xfer := mustLookupTransfer(mustApp())
	must(client.AppTransferDelete(xfer.Id))
	log.Printf("Canceled transfer of %s to %s.", xfer.App.Name, xfer.Recipient.Email)
}

func mustLookupTransfer(appname string) (xfer *heroku.AppTransfer) {
	// If the API starts allowing app identity instead of requiring
	// app-transfer UUID, this lookup will be unnecessary.
	transfers, err := client.AppTransferList(nil)
	must(err)
	for i := range transfers {
		if transfers[i].App.Name == appname {
			xfer = &transfers[i]
			break
		}
	}
	if xfer == nil {
		log.Println("No pending transfer for " + appname + ".")
		os.Exit(1)
	}
	return
}

func updateTransferState(transferId, newstate string) error {
	_, err := client.AppTransferUpdate(transferId, newstate)
	return err
}
