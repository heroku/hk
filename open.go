package main

import "os"

var cmdOpen = &Command{
	Run:      runOpen,
	Usage:    "open",
	NeedsApp: true,
	Category: "app",
	Short:    "open app in a web browser" + extra,
	Long:     `Open opens the app in a web browser. (Assumes cedar.)`,
}

func runOpen(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	app, err := client.AppInfo(mustApp())
	must(err)
	must(openURL(app.WebURL))
}
