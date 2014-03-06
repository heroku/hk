package main

import (
	"log"
	"os/exec"

	"github.com/bgentry/heroku-go"
)

var cmdCreate = &Command{
	Run:      runCreate,
	Usage:    "create [-r <region>] [<name>]",
	Category: "app",
	Short:    "create an app",
	Long: `
Create creates a new Heroku app. If <name> is not specified, the
app is created with a random haiku name.

Options:

    -r <region>  Heroku region to create app in
    <name>       optional name for the app

Examples:

    $ hk create
    Created dodging-samurai-42.

    $ hk create -r eu myapp
    Created myapp.
`,
}

var flagRegion string

func init() {
	cmdCreate.Flag.StringVar(&flagRegion, "r", "", "region name")
}

func runCreate(cmd *Command, args []string) {
	var opts heroku.AppCreateOpts
	if flagRegion != "" {
		opts.Region = &flagRegion
	}
	if len(args) > 0 {
		opts.Name = &args[0]
	}
	app, err := client.AppCreate(&opts)
	must(err)
	exec.Command("git", "remote", "add", "heroku", app.GitURL).Run()
	log.Printf("Created %s.", app.Name)
}
