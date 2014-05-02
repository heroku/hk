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
	cmdCreate.Flag.StringVarP(&flagRegion, "region", "r", "", "region name")
}

func runCreate(cmd *Command, args []string) {
	appname := ""
	if len(args) > 0 {
		appname = args[0]
	}

	// check for default org
	defaultOrgName := ""
	orgs, err := client.OrganizationList(nil)
	must(err)
	for _, org := range orgs {
		if org.Default {
			defaultOrgName = org.Name
			break
		}
	}

	if defaultOrgName == "" {
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
		log.Printf("Created %s.", app.Name)
	} else {
		var opts heroku.OrganizationAppCreateOpts
		if flagRegion != "" {
			opts.Region = &flagRegion
		}
		if appname != "" {
			opts.Name = &appname
		}

		app, err := client.OrganizationAppCreate(defaultOrgName, &opts)
		must(err)
		exec.Command("git", "remote", "add", "heroku", app.GitURL).Run()
		log.Printf("Created %s.", app.Name)
	}
}
