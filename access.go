package main

import (
	"log"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/bgentry/heroku-go"
)

var cmdAccess = &Command{
	Run:      runAccess,
	Usage:    "access",
	NeedsApp: true,
	Category: "access",
	Short:    "list access permissions" + extra,
	Long: `
List access permissions for an app. The owner is shown first, and
collaborators are then listed alphabetically.

Examples:

    $ hk access
    b@heroku.com    owner
    max@heroku.com  collaborator
`,
}

func runAccess(cmd *Command, args []string) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	if len(args) != 0 {
		cmd.PrintUsage()
		os.Exit(2)
	}

	// Org collaborators works for all apps and gives us exactly the data we need.
	orgCollaborators, err := client.OrganizationAppCollaboratorList(mustApp(), nil)
	must(err)

	sort.Sort(accessByRoleAndEmail(orgCollaborators))
	for _, oc := range orgCollaborators {
		listRec(w,
			oc.User.Email,
			oc.Role,
			prettyTime{oc.UpdatedAt},
		)
	}
}

type accessByRoleAndEmail []heroku.OrganizationAppCollaborator

func (a accessByRoleAndEmail) Len() int      { return len(a) }
func (a accessByRoleAndEmail) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a accessByRoleAndEmail) Less(i, j int) bool {
	return a[i].Role == "owner" || a[i].User.Email < a[j].User.Email
}

var cmdAccessAdd = &Command{
	Run:      runAccessAdd,
	Usage:    "access-add [-s] <email>",
	NeedsApp: true,
	Category: "access",
	Short:    "give a user access to an app" + extra,
	Long: `
Give another Heroku user access to an app.

Options:

    -s  add user silently with no email notification

Examples:

    $ hk access-add user@me.com
    Granted user@me.com access to myapp.

    $ hk access-add -s anotheruser@me.com
    Granted anotheruser@me.com access to myapp.
`,
}

var flagSilent bool

func init() {
	cmdAccessAdd.Flag.BoolVarP(&flagSilent, "silent", "s", false, "add user silently with no email notification")
}

func runAccessAdd(cmd *Command, args []string) {
	appname := mustApp()
	if len(args) != 1 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	user := args[0]

	var err error
	if isAppInOrg(mustGetOrgApp(appname)) {
		opts := heroku.OrganizationAppCollaboratorCreateOpts{Silent: &flagSilent}
		_, err = client.OrganizationAppCollaboratorCreate(appname, user, &opts)
	} else {
		opts := heroku.CollaboratorCreateOpts{Silent: &flagSilent}
		_, err = client.CollaboratorCreate(appname, user, &opts)
	}
	must(err)
	log.Printf("Granted %s access to %s.", user, appname)
}

var cmdAccessRemove = &Command{
	Run:      runAccessRemove,
	Usage:    "access-remove <email>",
	NeedsApp: true,
	Category: "access",
	Short:    "remove a user's access to an app" + extra,
	Long: `
Remove another Heroku user's access to an app.

Examples:

    $ hk access-remove user@me.com
    Removed user@me.com from access to myapp.
`,
}

func runAccessRemove(cmd *Command, args []string) {
	appname := mustApp()
	if len(args) != 1 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	user := args[0]

	if isAppInOrg(mustGetOrgApp(appname)) {
		must(client.OrganizationAppCollaboratorDelete(appname, user))
	} else {
		must(client.CollaboratorDelete(appname, user))
	}
	log.Printf("Removed %s from access to %s.", user, appname)
}
