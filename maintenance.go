package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bgentry/heroku-go"
)

var cmdMaintenance = &Command{
	Run:      runMaintenance,
	Usage:    "maintenance",
	NeedsApp: true,
	Category: "app",
	Short:    "show app maintenance mode" + extra,
	Long: `
Maintenance shows the current maintenance mode state of an app.

Example:

    $ hk maintenance
    enabled
`,
}

func runMaintenance(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	app, err := client.AppInfo(mustApp())
	must(err)
	state := "disabled"
	if app.Maintenance {
		state = "enabled"
	}
	fmt.Println(state)
}

var cmdMaintenanceEnable = &Command{
	Run:      runMaintenanceEnable,
	Usage:    "maintenance-enable",
	NeedsApp: true,
	Category: "app",
	Short:    "enable maintenance mode" + extra,
	Long: `
Enables maintenance mode on an app.

Example:

    $ hk maintenance-enable
    Enabled maintenance mode on myapp.
`,
}

func runMaintenanceEnable(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	newmode := true
	app, err := client.AppUpdate(mustApp(), &heroku.AppUpdateOpts{Maintenance: &newmode})
	must(err)
	log.Printf("Enabled maintenance mode on %s.", app.Name)
}

var cmdMaintenanceDisable = &Command{
	Run:      runMaintenanceDisable,
	Usage:    "maintenance-disable",
	NeedsApp: true,
	Category: "app",
	Short:    "disable maintenance mode" + extra,
	Long: `
Disables maintenance mode on an app.

Example:

    $ hk maintenance-disable
    Disabled maintenance mode on myapp.
`,
}

func runMaintenanceDisable(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	newmode := false
	app, err := client.AppUpdate(mustApp(), &heroku.AppUpdateOpts{Maintenance: &newmode})
	must(err)
	log.Printf("Disabled maintenance mode on %s.", app.Name)
}
