package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

var cmdLs = &Command{
	Run:   runLs,
	Usage: "ls [-l] [resource...]",
	Short: "list addons",
	Long: `
       hk ls [-l] addons [name...]

Command hk ls lists addons.

Options:

    -l       long listing

Long listing for addons shows the type of the addon, owner, name
of the resource, and the config var it's attached to.

Examples:

    $ hk ls addons
    DATABASE_URL
    REDIS_URL

    $ hk ls -l addons REDIS_URL
    redistogo:nano  me  soaring-ably-1234  REDIS_URL
`,
}

func init() {
	cmdLs.Flag.BoolVar(&flagLong, "l", false, "long listing")
}

func runLs(cmd *Command, args []string) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	list(w, cmd, args)
	w.Flush()
}

func list(w io.Writer, cmd *Command, args []string) {
	switch a0 := args[0]; {
	case strings.HasPrefix("addons", a0):
		listAddons(w, args[1:])
	}
}


func abbrevEmailResources(ms []*mergedAddon) {
	domains := make(map[string]int)
	for _, m := range ms {
		parts := strings.SplitN(m.Owner, "@", 2)
		if len(parts) == 2 {
			domains["@"+parts[1]]++
		}
	}
	smax, nmax := "", 0
	for s, n := range domains {
		if n > nmax {
			smax = s
			nmax = n
		}
	}
	for _, m := range ms {
		if strings.HasSuffix(m.Owner, smax) {
			m.Owner = m.Owner[:len(m.Owner)-len(smax)]
		}
	}
}

func listAddons(w io.Writer, names []string) {
	ms := getMergedAddons(mustApp())
	abbrevEmailResources(ms)
	for i, s := range names {
		names[i] = strings.ToLower(s)
	}
	for _, m := range ms {
		if len(names) == 0 || addonMatch(m, names) {
			listAddon(w, m)
		}
	}
}

func addonMatch(m *mergedAddon, a []string) bool {
	for _, s := range a {
		if s == strings.ToLower(m.Type) {
			return true
		}
		if s == strings.ToLower(m.ID) {
			return true
		}
	}
	return false
}

func listAddon(w io.Writer, m *mergedAddon) {
	if flagLong {
		listRec(w,
			m.Type,
			abbrev(m.Owner, 10),
			m.ID,
		)
	} else {
		fmt.Fprintln(w, m.String())
	}
}

func listRec(w io.Writer, a ...interface{}) {
	for i, x := range a {
		fmt.Fprint(w, x)
		if i+1 < len(a) {
			w.Write([]byte{'\t'})
		} else {
			w.Write([]byte{'\n'})
		}
	}
}
