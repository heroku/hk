package main

var cmdOpen = &Command{
	Run:      runOpen,
	Name:     "open",
	Usage:    "[-a <app>]",
	Category: "app",
	Short:    "open app in a web browser" + extra,
	Long:     `Open opens the app in a web browser. (Assumes cedar.)`,
}

func init() {
	cmdOpen.Flag.StringVar(&flagApp, "a", "", "app name")
}

func runOpen(cmd *Command, args []string) {
	must(openURL("https://" + mustApp() + ".herokuapp.com/"))
}
