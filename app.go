package main

import (
	"fmt"
)

var cmdApp = &Command{
	Run:      runApp,
	Usage:    "app",
	Category: "app",
	Short:    "show current directory app, if any" + extra,
	Long: `
App looks for a git remote named "heroku" with a remote URL in the
correct form. If successful, it prints the corresponding app name.
Otherwise, it prints an message to stderr and exits with a nonzero
status.

To suppress the error message, run 'hk app 2>/dev/null'.
`,
}

func runApp(cmd *Command, args []string) {
	fmt.Println(mustApp())
}
