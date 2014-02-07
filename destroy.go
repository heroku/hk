package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/heroku/hk/term"
	"github.com/mgutz/ansi"
)

var cmdDestroy = &Command{
	Run:      runDestroy,
	Usage:    "destroy <name>",
	Category: "app",
	Short:    "destroy an app",
	Long: `
Destroy destroys a heroku app. There is no going back, so be
sure you mean it. The command will prompt for confirmation, or
accept confirmation via stdin.

Example:

    $ hk destroy myapp
    warning: This will destroy myapp and its add-ons. Please type "myapp" to continue:
    Destroyed myapp.

    $ echo myapp | hk destroy myapp
    Destroyed myapp.
`,
}

func runDestroy(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	appname := args[0]

	if term.IsTerminal(os.Stdin) {
		printWarning("This will destroy %s and its add-ons. Please type %q to continue:", appname, appname)
		fmt.Printf(ansi.Color("> ", "+b") + ansi.ColorCode("reset"))
	}
	var confirm string
	if _, err := fmt.Scanln(&confirm); err != nil {
		printFatal(err.Error())
	}

	if confirm != appname {
		printFatal("Confirmation did not match %q.", appname)
	}

	must(client.AppDelete(appname))
	log.Printf("Destroyed %s.", appname)
	remotes, _ := gitRemotes()
	for remote, remoteApp := range remotes {
		if appname == remoteApp {
			exec.Command("git", "remote", "rm", remote).Run()
		}
	}
}
