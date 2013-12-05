package main

import (
	"os"
	"os/exec"
)

var cmdDestroy = &Command{
	Run:   runDestroy,
	Usage: "destroy <name>",
	Short: "destroy an app",
	Long: `
Destroy destroys a heroku app.

There is no going back, so be sure you mean it.

Example:

    $ hk destroy myapp
`,
}

func runDestroy(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	name := args[0]
	must(client.AppDelete(name))
	for _, remote := range gitRemotes(gitURL(name)) {
		exec.Command("git", "remote", "rm", remote).Run()
	}
}
