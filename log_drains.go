package main

import (
	"log"
	"os"
	"text/tabwriter"
)

var cmdLogDrains = &Command{
	Run:      runLogDrains,
	Usage:    "log-drains",
	NeedsApp: true,
	Category: "app",
	Short:    "list log drains" + extra,
	Long: `
Lists log drains on an app.

Example:

    $ hk log-drains
    7f89b6bb-08af-4343-b0b4-d0415dd81712  d.b5f66703-6fb2-4195-a4b1-3ab2f1e3423f  syslog://my.log.host
    23fcdb8a-3095-46f5-abc2-c5f293c54cf1  d.3f17356d-3b5d-4de2-a6aa-cc367e4a8fc8  syslog://my.other.log.host
`,
}

func runLogDrains(cmd *Command, args []string) {
	drains, err := client.LogDrainList(mustApp(), nil)
	must(err)

	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	for _, drain := range drains {
		listRec(w, drain.Id, drain.Token, drain.URL)
	}
}

var cmdLogDrainAdd = &Command{
	Run:      runLogDrainAdd,
	Usage:    "log-drain-add <url>",
	NeedsApp: true,
	Category: "app",
	Short:    "add a log drain" + extra,
	Long: `
Adds a log drain to an app.

Example:

    $ hk log-drain-add syslog://my.log.host
    Added log drain to myapp.
`,
}

func runLogDrainAdd(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}

	url := args[0]
	_, err := client.LogDrainCreate(mustApp(), url)
	must(err)
	log.Printf("Added log drain to %s.", mustApp())
}

var cmdLogDrainRemove = &Command{
	Run:      runLogDrainRemove,
	Usage:    "log-drain-remove <id or url>",
	NeedsApp: true,
	Category: "app",
	Short:    "remove a log drain" + extra,
	Long: `
Removes a log drain from an app.

Example:

    $ hk log-drain-remove 7f89b6bb-08af-4343-b0b4-d0415dd81712
    Removed log drain from myapp.

    $ hk log-drain-remove syslog://my.log.host
    Removed log drain from myapp.
`,
}

func runLogDrainRemove(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}

	drainId := args[0]
	must(client.LogDrainDelete(mustApp(), drainId))
	log.Printf("Removed log drain from %s.", mustApp())
}
