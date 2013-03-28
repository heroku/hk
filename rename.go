package main

import (
	"fmt"
	"os"
)

var cmdRename = &Command{
	Run:   runRename,
	Usage: "rename old new",
	Short: "rename an app",
	Long: `
Rename renames a heroku app.

Example:

  $ hk rename myapp myapp2
`,
}

func runRename(cmd *Command, args []string) {
	if len(args) != 2 {
		cmd.printUsage()
		os.Exit(2)
	}
	oldname, newname := args[0], args[1]
	var app App
	p := map[string]string{"name": newname}
	must(Put(&app, "/apps/"+oldname, p))
	fmt.Println("Renamed app to: " + app.Name)
	fmt.Println("Ensure you update your git remote URL.")
	// should we automatically update the remote if they specify an app
	// or via mustApp + conditional logic - RM
}
