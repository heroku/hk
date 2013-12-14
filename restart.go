package main

import (
	"log"
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
  $ hk restart web
  $ hk restart web.1
`,
}

func runRestart(cmd *Command, args []string) {
	if len(args) > 1 {
		log.Fatal("Invalid usage. See 'hk help restart'")
	}

	if len(args) == 1 {
		must(client.DynoRestart(mustApp(), args[0]))
	} else {
		must(client.DynoRestartAll(mustApp()))
	}
}
