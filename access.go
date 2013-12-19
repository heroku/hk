package main

import (
	"log"
	"os"
	"sort"
	"text/tabwriter"
	"time"

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

func runAccess(cmd *Command, names []string) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	ma := getMergedAccess(mustApp())
	for _, m := range ma {
		listRec(w,
			m.User,
			m.Role,
			prettyTime{m.Time},
		)
	}
}

type mergedAccess struct {
	User string
	Role string
	Time time.Time
}

func getMergedAccess(appname string) []*mergedAccess {
	var collaborators []heroku.Collaborator
	var app *heroku.App
	ch := make(chan error)
	go func() {
		var err error
		collaborators, err = client.CollaboratorList(appname, nil)
		ch <- err
	}()
	go func() {
		var err error
		app, err = client.AppInfo(appname)
		ch <- err
	}()
	if err := <-ch; err != nil {
		log.Fatal(err)
	}
	if err := <-ch; err != nil {
		log.Fatal(err)
	}
	return mergeAccess(app, collaborators)
}

type accessByRoleAndUser []*mergedAccess

func (a accessByRoleAndUser) Len() int      { return len(a) }
func (a accessByRoleAndUser) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a accessByRoleAndUser) Less(i, j int) bool {
	return a[i].Role == "owner" || a[i].User < a[j].User
}

func mergeAccess(app *heroku.App, collaborators []heroku.Collaborator) (ma []*mergedAccess) {
	// User, Role, Time
	for _, c := range collaborators {
		role := "collaborator"
		if app.Owner.Email == c.User.Email {
			role = "owner"
		}
		m := &mergedAccess{
			User: c.User.Email,
			Role: role,
			Time: c.UpdatedAt,
		}
		ma = append(ma, m)
	}
	sort.Sort(accessByRoleAndUser(ma))
	return ma
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

    $ hk access-add -s anotheruser@me.com
`,
}

var flagSilent bool

func init() {
	cmdAccessAdd.Flag.BoolVar(&flagSilent, "s", false, "add user silently with no email notification")
}

func runAccessAdd(cmd *Command, args []string) {
	if len(args) < 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	opts := heroku.CollaboratorCreateOpts{Silent: &flagSilent}
	_, err := client.CollaboratorCreate(mustApp(), args[0], &opts)
	must(err)
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
`,
}

func runAccessRemove(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	must(client.CollaboratorDelete(mustApp(), args[0]))
}
