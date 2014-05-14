package main

import (
	"log"
	"os/exec"

	"github.com/bgentry/heroku-go"
)

var cmdCreate = &Command{
	Run:      runCreate,
	Usage:    "create [-r <region>] [-o <org>] [<name>]",
	Category: "app",
	Short:    "create an app",
	Long: `
Create creates a new Heroku app. If <name> is not specified, the
app is created with a random haiku name.

Options:

    -r <region>  Heroku region to create app in
    -o <org>     name of Heroku organization to create app in
    <name>       optional name for the app

Examples:

    $ hk create
    Created dodging-samurai-42.

    $ hk create -r eu myapp
    Created myapp.
`,
}

var flagRegion string
var flagOrgName string

func init() {
	cmdCreate.Flag.StringVarP(&flagRegion, "region", "r", "", "region name")
	cmdCreate.Flag.StringVarP(&flagOrgName, "org", "o", "", "organization name")
}

func runCreate(cmd *Command, args []string) {
	appname := ""
	if len(args) > 0 {
		appname = args[0]
	}

	if flagOrgName == "personal" { // "personal" means "no org"
		var opts heroku.AppCreateOpts
		if flagRegion != "" {
			opts.Region = &flagRegion
		}
		if appname != "" {
			opts.Name = &appname
		}

		app, err := client.AppCreate(&opts)
		must(err)
		exec.Command("git", "remote", "add", "heroku", app.GitURL).Run()
		printCreateSuccess(fromApp(*app))
		return
	}

	var opts heroku.OrganizationAppCreateOpts
	if appname != "" {
		opts.Name = &appname
	}
	if flagOrgName != "" {
		opts.Organization = &flagOrgName
	}
	if flagRegion != "" {
		opts.Region = &flagRegion
	}

	app, err := client.OrganizationAppCreate(&opts)
	must(err)
	exec.Command("git", "remote", "add", "heroku", app.GitURL).Run()
	printCreateSuccess(fromOrgApp(*app))
}

func printCreateSuccess(app hkapp) {
	if app.Organization != "" {
		log.Printf("Created %s in the %s org.", app.Name, app.Organization)
	} else {
		log.Printf("Created %s.", app.Name)
	}
}
