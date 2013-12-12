package main

import (
	"fmt"
	"github.com/bgentry/heroku-go"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
)

var cmdAddons = &Command{
	Run:      runAddons,
	Usage:    "addons [<provider>:<plan>...]",
	NeedsApp: true,
	Category: "add-on",
	Short:    "list addons",
	Long: `
Lists addons.

Examples:

    $ hk addons
    heroku-postgresql:crane
    pgbackups:plus

    $ hk addons pgbackups:plus
    pgbackups:plus
`,
}

func runAddons(cmd *Command, names []string) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()
	listAddons(w, names)
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
		if s == strings.ToLower(m.Id) {
			return true
		}
	}
	return false
}

func listAddon(w io.Writer, m *mergedAddon) {
	fmt.Fprintln(w, m.String())
}

type mergedAddon struct {
	Type  string
	Owner string
	Id    string
}

func (m *mergedAddon) String() string {
	return m.Type
}

func getMergedAddons(appname string) []*mergedAddon {
	var addons []heroku.Addon
	app := new(heroku.App)
	app.Name = appname
	ch := make(chan error)
	go func() {
		var err error
		addons, err = client.AddonList(app.Name, nil)
		ch <- err
	}()
	go func() {
		var err error
		app, err = client.AppInfo(app.Name)
		ch <- err
	}()
	if err := <-ch; err != nil {
		log.Fatal(err)
	}
	if err := <-ch; err != nil {
		log.Fatal(err)
	}
	return mergeAddons(app, addons)
}

func mergeAddons(app *heroku.App, addons []heroku.Addon) (ms []*mergedAddon) {
	// Type, Owner, Id
	for _, a := range addons {
		m := new(mergedAddon)
		ms = append(ms, m)
		m.Type = a.Plan.Name
		m.Owner = app.Owner.Email
		m.Id = a.Id
	}

	sort.Sort(mergedAddonsByType(ms))
	return ms
}

type mergedAddonsByType []*mergedAddon

func (a mergedAddonsByType) Len() int           { return len(a) }
func (a mergedAddonsByType) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a mergedAddonsByType) Less(i, j int) bool { return a[i].Type < a[j].Type }

var cmdAddonAdd = &Command{
	Run:      runAddonAdd,
	Usage:    "addon-add <provider>[:<plan>] [<config>=<value>...]",
	NeedsApp: true,
	Category: "add-on",
	Short:    "add an addon",
	Long: `
Adds an addon to an app.

Examples:

    $ hk addon-add heroku-postgresql:hobby-basic

    $ hk addon-add heroku-postgresql:standard-tengu
`,
}

func runAddonAdd(cmd *Command, args []string) {
	if len(args) == 0 {
		cmd.printUsage()
		os.Exit(2)
	}
	plan := args[0]
	if strings.IndexRune(plan, ':') == -1 {
		// has provider name, but missing plan name
		cmd.printUsage()
		os.Exit(2)
	}
	var opts heroku.AddonCreateOpts
	if len(args) > 1 {
		config, err := parseAddonAddConfig(args[1:])
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
		opts = heroku.AddonCreateOpts{Config: config}
	}
	_, err := client.AddonCreate(mustApp(), plan, opts)
	must(err)
}

func parseAddonAddConfig(config []string) (*map[string]string, error) {
	conf := make(map[string]string, len(config))
	for _, kv := range config {
		iEq := strings.IndexRune(kv, '=')
		if iEq < 1 || len(kv) < iEq+2 {
			return nil, fmt.Errorf("Invalid option '%s', must be of form 'key=value'", kv)
		}
		val := kv[iEq+1:]
		if val[0] == '\'' {
			val = strings.Trim(val, "'")
		} else if val[0] == '"' {
			val = strings.Trim(val, "\"")
		}
		conf[kv[:iEq]] = val
	}
	return &conf, nil
}

var cmdAddonRemove = &Command{
	Run:      runAddonRemove,
	Usage:    "addon-remove <provider>:<plan>",
	NeedsApp: true,
	Category: "add-on",
	Short:    "remove an addon",
	Long: `
Removes an addon from an app.

Examples:

    $ hk addon-remove heroku-postgresql:basic-dev

    $ hk addon-remove heroku-postgresql:standard-tengu
`,
}

func runAddonRemove(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	plan := args[0]
	if strings.IndexRune(plan, ':') == -1 {
		// has provider name, but missing plan name
		cmd.printUsage()
		os.Exit(2)
	}
	err := client.AddonDelete(mustApp(), plan)
	must(err)
}

var cmdAddonOpen = &Command{
	Run:      runAddonOpen,
	Usage:    "addon-open <provider>[:<plan>]",
	NeedsApp: true,
	Category: "add-on",
	Short:    "open an addon" + extra,
	Long: `
Open the addon's management page in your default web browser.
`,
}

func runAddonOpen(cmd *Command, args []string) {
	app := mustApp()
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	must(openURL("https://addons-sso.heroku.com/apps/" + app + "/addons/" + args[0]))
}
