package main

import (
	"fmt"
)

var nsApp = &Namespace{
	Name: "app",
	Commands: []*Command{
		cmdAppList,
		cmdAppCreate,
		cmdAppInfo,
		cmdAppRename,
		cmdAppDestroy,
		cmdAppCurrent,
	},
	Short: "manage apps",
}

var cmdAppCurrent = &Command{
	Run:   runAppCurrent,
	Usage: "current",
	Short: "show current directory app, if any" + extra,
	Long: `
Current looks for a git remote named "heroku" with a remote URL in the
correct form. If successful, it prints the corresponding app name.
Otherwise, it prints an message to stderr and exits with a nonzero
status.

To suppress the error message, run 'hk app current 2>/dev/null'.
`,
}

func runAppCurrent(cmd *Command, args []string) {
	fmt.Println(mustApp())
}
