package main

import (
	"log"
	"os"
	"os/exec"
)

var cmdDestroy = &Command{
	Run:      runDestroy,
	Usage:    "destroy <name>",
	Category: "app",
	Short:    "destroy an app",
	Long: `
Destroy destroys a heroku app.

There is no going back, so be sure you mean it.

Example:

    $ hk destroy myapp
    Destroyed myapp.
`,
}

func runDestroy(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	appname := args[0]
	must(client.AppDelete(appname))
	log.Printf("Destroyed %s.", appname)
	remotes, _ := gitRemotes()
	for remote, remoteApp := range remotes {
		if appname == remoteApp {
			exec.Command("git", "remote", "rm", remote).Run()
		}
	}
}
