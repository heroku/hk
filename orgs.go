package main

import (
	"io"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/bgentry/heroku-go"
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

	orgs, err := client.OrganizationList(nil)
	must(err)
	printOrgsList(w, orgs)
}

func printOrgsList(w io.Writer, orgs []heroku.Organization) {
	sort.Sort(orgsByName(orgs))
	for _, org := range orgs {
		listRec(w,
			org.Name,
			org.Role,
		)
	}
}

type orgsByName []heroku.Organization

func (o orgsByName) Len() int           { return len(o) }
func (o orgsByName) Swap(i, j int)      { o[i], o[j] = o[j], o[i] }
func (o orgsByName) Less(i, j int) bool { return o[i].Name < o[j].Name }

// Returns true if the app is in an org, and false otherwise.
func isAppInOrg(app *heroku.OrganizationApp) bool {
	return app.Organization != nil
}

// This function uses must(err), so the program will exit if an error is
// encountered.
func mustGetOrgApp(appname string) *heroku.OrganizationApp {
	app, err := client.OrganizationAppInfo(appname)
	must(err)
	return app
}
