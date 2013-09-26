package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
)

var cmdAddons = &Command{
	Run:   runAddons,
	Usage: "addons [-l] [resource...]",
	Short: "list addons",
	Long: `
Lists addons.

Options:

		-l       long listing

Long listing shows the type of the addon, owner, name of the
resource, and the config var it's attached to.

Examples:

		$ hk ls addons
		DATABASE_URL
		REDIS_URL

		$ hk ls -l addons REDIS_URL
		redistogo:nano  me  soaring-ably-1234  REDIS_URL
`,
}

func init() {
	cmdAddons.Flag.BoolVar(&flagLong, "l", false, "long listing")
}

func runAddons(cmd *Command, names []string) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	listAddons(w, names)
	w.Flush()
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

type mergedAddon struct {
	Type  string
	Owner string
	ID    string
}

func (m *mergedAddon) String() string {
	return m.Type
}

func getMergedAddons(appname string) []*mergedAddon {
	var addons []*Addon
	app := new(App)
	app.Name = mustApp()
	ch := make(chan error)
	go func() { ch <- Get(&addons, "/apps/"+app.Name+"/addons") }()
	go func() { ch <- Get(app, "/apps/"+app.Name) }()
	if err := <-ch; err != nil {
		log.Fatal(err)
	}
	if err := <-ch; err != nil {
		log.Fatal(err)
	}
	return mergeAddons(app, addons)
}

func mergeAddons(app *App, addons []*Addon) (ms []*mergedAddon) {
	// Type, Name, Owner
	for _, a := range addons {
		m := new(mergedAddon)
		ms = append(ms, m)
		m.Type = a.Plan.Name
		m.Owner = app.Owner.Email
		m.ID = a.ID
	}

	sort.Sort(mergedAddonsByType(ms))
	return ms
}

type mergedAddonsByType []*mergedAddon

func (a mergedAddonsByType) Len() int           { return len(a) }
func (a mergedAddonsByType) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a mergedAddonsByType) Less(i, j int) bool { return a[i].Type < a[j].Type }
