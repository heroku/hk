package main

import (
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
