package main

import (
	"log"
)

var cmdRestart = &Command{
	Run:   runRestart,
	Usage: "restart [type or name]",
	Short: "restart dynos",
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

	path := "/apps/" + mustApp() + "/dynos"

	if len(args) == 1 {
		path += "/" + args[0]
	}

	must(Delete(path))
}
