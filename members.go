package main

import (
	"log"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/bgentry/heroku-go"
)

var cmdMembers = &Command{
	Run:      runMembers,
	Usage:    "members <orgname>",
	NeedsApp: false,
	Category: "members",
	Short:    "list member permissions for an organization" + extra,
	Long: `
List member permissions for an organization. Members are listed
alphabetically.

Examples:

    $ hk members
    b@heroku.com    member
    max@heroku.com  admin
`,
}

func runMembers(cmd *Command, args []string) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	if len(args) != 1 {
		cmd.PrintUsage()
		os.Exit(2)
	}

	orgname := args[0]

	orgMembers, err := client.OrganizationMemberList(orgname, nil)
	must(err)

	sort.Sort(membersByEmail(orgMembers))
	for _, oc := range orgMembers {
		listRec(w,
			oc.Email,
			oc.Role,
		)
	}
}

type membersByEmail []heroku.OrganizationMember

func (a membersByEmail) Len() int      { return len(a) }
func (a membersByEmail) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a membersByEmail) Less(i, j int) bool {
	return a[i].Email < a[j].Email
}

var role string

var cmdMemberAdd = &Command{
	Run:      runMemberAdd,
	Usage:    "member-add [-r <role>] <orgname> <email>",
	NeedsApp: false,
	Category: "members",
	Short:    "add a member to an organization" + extra,
	Long: `
Make another Heroku user to an organization. If no role
is provided, the command defaults to adding the user with
the role "member".

Options:
    -r <role>  Organization role to assign member

Examples:

    $ hk member-add myorg user@me.com
    Added user@me.com to myorg with role 'member'.

    $ hk member-add -r admin myorg anotheruser@me.com
    Added anotheruser@me.com to myorg with role 'admin'.
`,
}

func init() {
	cmdMemberAdd.Flag.StringVarP(&role, "role", "r", "member", "role name")
}

func runMemberAdd(cmd *Command, args []string) {
	if len(args) != 2 {
		cmd.PrintUsage()
		os.Exit(2)
	}

	orgname, user := args[0], args[1]

	_, err := client.OrganizationMemberCreateOrUpdate(orgname, user, role)
	must(err)

	log.Printf("Added %s to %s with role '%s'.", user, orgname, role)
}
