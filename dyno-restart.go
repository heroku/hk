package main

import (
	"log"
)

var cmdDynoRestart = &Command{
	Run:      runDynoRestart,
	NeedsApp: true,
	Usage:    "restart [type or name]",
	Short:    "restart dynos",
	Long: `
Restart all app dynos, all dynos of a specific type, or a single dyno.

Examples:

  $ hk dyno restart
  $ hk dyno restart web
  $ hk dyno restart web.1
`,
}

func runDynoRestart(cmd *Command, args []string) {
	if len(args) > 1 {
		log.Fatal("Invalid usage. See 'hk help dyno restart'")
	}

	if len(args) == 1 {
		must(client.DynoRestart(mustApp(), args[0]))
	} else {
		must(client.DynoRestartAll(mustApp()))
	}
}
