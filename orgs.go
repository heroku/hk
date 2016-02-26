package main

import (
	"io"
	"os"
	"text/tabwriter"

	"github.com/heroku/hk/Godeps/_workspace/src/github.com/bgentry/heroku-go"
)

var cmdOrgs = &Command{
	Run:      runOrgs,
	Usage:    "orgs",
	Category: "orgs",
	Short:    "list Heroku orgs",
	Long:     "Lists Heroku organizations that the user belongs to.",
}

func runOrgs(cmd *Command, args []string) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	orgs, err := client.OrganizationList(&heroku.ListRange{Field: "name"})
	must(err)
	printOrgsList(w, orgs)
}

func printOrgsList(w io.Writer, orgs []heroku.Organization) {
	for _, org := range orgs {
		listRec(w,
			org.Name,
			org.Role,
		)
	}
}

// Returns true if the app is in an org, and false otherwise.
func isAppInOrg(app *heroku.OrganizationApp) bool {
	return app != nil && app.Organization != nil
}

// This function uses must(err), so the program will exit if an error is
// encountered.
func mustGetOrgApp(appname string) *heroku.OrganizationApp {
	app, err := client.OrganizationAppInfo(appname)
	if err != nil {
		// the organization app info endpoint responds with a 404 if
		e, ok := err.(heroku.Error)
		if ok && e.Id == "not_found" {
			return nil
		}
		must(err)
	}
	return app
}
