package main

import (
	"os/exec"
)

var cmdOpen = &Command{
	Run:      runOpen,
	NeedsApp: true,
	Usage:    "open",
	Short:    "open app in a web browser",
	Long:     `Open opens the app in a web browser. (Assumes cedar.)`,
}

func runOpen(cmd *Command, args []string) {
	u := "https://" + mustApp() + ".herokuapp.com/"
	command := "open"
	if _, err := exec.LookPath("xdg-open"); err == nil {
		command = "xdg-open"
	}
	exec.Command(command, u).Start()
}
