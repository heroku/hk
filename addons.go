package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/bgentry/heroku-go"
)

var cmdAddons = &Command{
	Run:      runAddons,
	Usage:    "addons [<service>:<plan>...]",
	NeedsApp: true,
	Category: "add-on",
	Short:    "list addons",
	Long: `
Lists addons.

Examples:

    $ hk addons
    heroku-postgresql-blue  heroku-postgresql:crane  Nov 19 12:40
    pgbackups               pgbackups:plus           Sep 30 15:43

    $ hk addons pgbackups
    pgbackups  pgbackups:plus  Sep 30 15:43
`,
}

func runAddons(cmd *Command, names []string) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	appname := mustApp()
	addons, err := client.AddonList(appname, nil)
	if err != nil {
		printError(err.Error())
	}
	for i, s := range names {
		names[i] = strings.ToLower(s)
	}
	for _, a := range addons {
		if len(names) == 0 || addonMatch(a, names) {
			listAddon(w, a)
		}
	}
}

func addonMatch(a heroku.Addon, names []string) bool {
	for _, name := range names {
		if name == strings.ToLower(a.Name) {
			return true
		}
		if name == strings.ToLower(a.Plan.Name) {
			return true
		}
		if name == strings.ToLower(a.Id) {
			return true
		}
	}
	return false
}

func listAddon(w io.Writer, a heroku.Addon) {
	name := a.Name
	if name == "" {
		name = "[unnamed]"
	}
	listRec(w,
		name,
		a.Plan.Name,
		prettyTime{a.CreatedAt},
	)
}

var cmdAddonAdd = &Command{
	Run:      runAddonAdd,
	Usage:    "addon-add <service>[:<plan>] [<config>=<value>...]",
	NeedsApp: true,
	Category: "add-on",
	Short:    "add an addon",
	Long: `
Adds an addon to an app.

Examples:

    $ hk addon-add heroku-postgresql
    Added heroku-postgresql:hobby-dev to myapp.

    $ hk addon-add heroku-postgresql:standard-tengu
    Added heroku-postgresql:standard-tengu to myapp.
`,
}

func runAddonAdd(cmd *Command, args []string) {
	appname := mustApp()
	if len(args) == 0 {
		cmd.printUsage()
		os.Exit(2)
	}
	plan := args[0]
	var opts heroku.AddonCreateOpts
	if len(args) > 1 {
		config, err := parseAddonAddConfig(args[1:])
		if err != nil {
			log.Println(err)
			os.Exit(2)
		}
		opts = heroku.AddonCreateOpts{Config: config}
	}
	addon, err := client.AddonCreate(appname, plan, &opts)
	must(err)
	log.Printf("Added %s to %s.", addon.Plan.Name, appname)
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
	Usage:    "addon-remove <name>",
	NeedsApp: true,
	Category: "add-on",
	Short:    "remove an addon",
	Long: `
Removes an addon from an app.

Examples:

    $ hk addon-remove heroku-postgresql-blue
    Removed heroku-postgresql-blue from myapp.

    $ hk addon-remove redistogo
    Removed redistogo from myapp.
`,
}

func runAddonRemove(cmd *Command, args []string) {
	appname := mustApp()
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	name := args[0]
	if strings.IndexRune(name, ':') != -1 {
		// specified an addon with plan name, unsupported in v3
		log.Println("Please specify an addon name, not a plan name.")
		cmd.printUsage()
		os.Exit(2)
	}
	checkAddonError(client.AddonDelete(appname, name))
	log.Printf("Removed %s from %s.", name, appname)
}

var cmdAddonOpen = &Command{
	Run:      runAddonOpen,
	Usage:    "addon-open <name>",
	NeedsApp: true,
	Category: "add-on",
	Short:    "open an addon" + extra,
	Long: `
Open the addon's management page in your default web browser.

Examples:

    $ hk addon-open heroku-postgresql-blue

    $ hk addon-open redistogo
`,
}

func runAddonOpen(cmd *Command, args []string) {
	appname := mustApp()
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	name := args[0]
	// look up addon to make sure it exists and to get plan name
	a, err := client.AddonInfo(appname, name)
	checkAddonError(err)
	must(openURL("https://addons-sso.heroku.com/apps/" + appname + "/addons/" + a.Plan.Name))
}

func checkAddonError(err error) {
	if err != nil {
		if hkerr, ok := err.(heroku.Error); ok && hkerr.Id == "not_found" {
			log.Println(err, "Choose an addon name from `hk addons`.")
		} else {
			log.Println(err)
		}
		os.Exit(2)
	}
}
