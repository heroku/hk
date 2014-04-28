package main

import (
	"log"
	"os"
	"strings"
)

var cmdRestart = &Command{
	Run:      runRestart,
	Usage:    "restart [<type or name>]",
	NeedsApp: true,
	Category: "dyno",
	Short:    "restart dynos",
	Long: `
Restart all app dynos, all dynos of a specific type, or a single dyno.

Examples:

    $ hk restart
    Restarted all dynos on myapp.

    $ hk restart web
    Restarted web dynos on myapp.

    $ hk restart web.1
    Restarted web.1 dyno on myapp.
`,
}

func runRestart(cmd *Command, args []string) {
	appname := mustApp()
	if len(args) > 1 {
		cmd.PrintUsage()
		os.Exit(2)
	}

	target := "all"
	if len(args) == 1 {
		target = args[0]
		must(client.DynoRestart(appname, target))
	} else {
		must(client.DynoRestartAll(appname))
	}

	switch {
	case strings.Contains(target, "."):
		log.Printf("Restarted %s dyno for %s.", target, appname)
	default:
		log.Printf("Restarted %s dynos for %s.", target, appname)
	}
}
