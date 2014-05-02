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
    -o <org>     Name of Heroku organization to create app in
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

	orgName := ""
	// "personal" means "no org", skip the org lookup stuff
	if flagOrgName != "personal" {
		orgs, err := client.OrganizationList(nil)
		must(err)
		for _, org := range orgs {
			if flagOrgName != "" {
				// match org in orgs list
				if org.Name == flagOrgName {
					orgName = org.Name
				}
			} else {
				// check for default org
				if org.Default {
					orgName = org.Name
					break
				}
			}
		}
		if flagOrgName != "" && flagOrgName != orgName {
			// flagOrgName was provided but not found in orgs list
			printFatal("Heroku organization %s not found", flagOrgName)
		}
	}

	if orgName == "" {
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

		app, err := client.OrganizationAppCreate(orgName, &opts)
		must(err)
		exec.Command("git", "remote", "add", "heroku", app.GitURL).Run()
		log.Printf("Created %s.", app.Name)
	}
}
