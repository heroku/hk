package main

import (
	"fmt"
	"github.com/bgentry/heroku-go"
	"os"
)

var cmdAppRename = &Command{
	Run:   runAppRename,
	Usage: "rename <old> <new>",
	Short: "rename an app",
	Long: `
Rename renames a heroku app.

Example:

  $ hk app rename myapp myapp2
`,
}

func runAppRename(cmd *Command, args []string) {
	if len(args) != 2 {
		cmd.printUsage()
		os.Exit(2)
	}
	oldname, newname := args[0], args[1]
	app, err := client.AppUpdate(oldname, heroku.AppUpdateOpts{Name: &newname})
	must(err)
	fmt.Println("Renamed app to: " + app.Name)
	fmt.Println("Ensure you update your git remote URL.")
	// should we automatically update the remote if they specify an app
	// or via mustApp + conditional logic - RM
}
