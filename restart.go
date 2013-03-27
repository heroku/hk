package main

import (
	"log"
	"net/url"
	"strings"
)

func init() {
	cmdRestart.Flag.StringVar(&flagApp, "a", "", "app")
}

var cmdRestart = &Command{
	Run:   runRestart,
	Usage: "restart [-a app] [type or name]",
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

	v := make(url.Values)

	if len(args) == 1 {
		if strings.Index(args[0], ".") > 0 {
			v.Add("ps", args[0])
		} else {
			v.Add("type", args[0])
		}
	}

	must(Post(nil, "/apps/"+mustApp()+"/ps/restart", v))
}
