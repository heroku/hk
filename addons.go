package main

import (
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"sort"
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
	appname := mustApp()
	addons, err := client.AddonList(appname, nil)
	must(err)

	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

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
    Added heroku-postgresql:hobby-dev to myapp as heroku-postgresql-yellow.

    $ hk addon-add heroku-postgresql:standard-tengu
    Added heroku-postgresql:standard-tengu to myapp as heroku-postgresql-orange.
`,
}

func runAddonAdd(cmd *Command, args []string) {
	appname := mustApp()
	if len(args) == 0 {
		cmd.PrintUsage()
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
		// if this is a postgres addon, resolve fork/follow/rollback args
		provider, _ := splitProviderAndPlan(plan)
		if provider == hpgAddonName() && config != nil {
			for k := range *config {
				if i := stringsIndex(hpgOptNames, k); i != -1 {
					// contains an hpgOptNames key, we need to resolve these against envs
					appEnv, err := client.ConfigVarInfo(appname)
					must(err)
					must(hpgAddonOptResolve(config, appEnv))
					break
				}
			}
		}
		opts = heroku.AddonCreateOpts{Config: config}
	}
	addon, err := client.AddonCreate(appname, plan, &opts)
	must(err)
	log.Printf("Added %s to %s as %s.", addon.Plan.Name, appname, addon.Name)
}

func splitProviderAndPlan(providerAndPlan string) (provider string, plan string) {
	parts := strings.Split(providerAndPlan, ":")
	if len(parts) > 0 {
		provider = parts[0]
	}
	if len(parts) > 1 {
		plan = parts[1]
	}
	return
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

var cmdAddonDestroy = &Command{
	Run:      runAddonDestroy,
	Usage:    "addon-destroy <name>",
	NeedsApp: true,
	Category: "add-on",
	Short:    "destroy an addon",
	Long: `
Removes an addon from an app, permanently destroying any data
stored by that addon. The command will prompt for confirmation,
or accept confirmation via stdin.

Examples:

    $ hk addon-destroy heroku-postgresql-blue
    warning: This will destroy heroku-postgresql-blue on myapp. Please type "myapp" to continue:
    > myapp
    Destroyed heroku-postgresql-blue on myapp.

    $ echo myapp | hk addon-destroy redistogo
    Destroyed redistogo on myapp.
`,
}

func runAddonDestroy(cmd *Command, args []string) {
	appname := mustApp()
	if len(args) != 1 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	name := args[0]
	if strings.IndexRune(name, ':') != -1 {
		// specified an addon with plan name, unsupported in v3
		log.Println("Please specify an addon name, not a plan name.")
		cmd.PrintUsage()
		os.Exit(2)
	}

	warning := "This will destroy %s on %s. Please type %q to continue:"
	mustConfirm(fmt.Sprintf(warning, name, appname, appname), appname)

	checkAddonError(client.AddonDelete(appname, name))
	log.Printf("Destroyed %s on %s.", name, appname)
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
		cmd.PrintUsage()
		os.Exit(2)
	}
	name := args[0]
	// look up addon to make sure it exists and to get plan name
	a, err := client.AddonInfo(appname, name)
	checkAddonError(err)
	must(openURL("https://addons-sso.heroku.com/apps/" + appname + "/addons/" + a.Plan.Name))
}

var cmdAddonPlan = &Command{
	Run:      runAddonPlan,
	Usage:    "addon-plan <name> <plan>",
	NeedsApp: true,
	Category: "add-on",
	Short:    "change an addon's plan" + extra,
	Long: `
Change an addon's plan. Not all add-on providers support this

Examples:

    $ hk addon-plan redistogo small
    Changed redistogo plan to small on myapp.
`,
}

func runAddonPlan(cmd *Command, args []string) {
	appname := mustApp()
	if len(args) != 2 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	name := args[0]
	plan := args[1]

	addon, err := client.AddonInfo(appname, name)
	checkAddonError(err)

	// assemble service:plan string
	serviceAndPlan := strings.Split(addon.Plan.Name, ":")[0] + ":" + plan

	a, err := client.AddonUpdate(appname, name, serviceAndPlan)
	checkAddonError(err)
	log.Printf("Changed %s plan to %s on %s.", a.Name, plan, appname)
}

func checkAddonError(err error) {
	if err != nil {
		if hkerr, ok := err.(heroku.Error); ok && hkerr.Id == "not_found" {
			printFatal(err.Error() + " Choose an addon name from `hk addons`.")
		} else {
			printFatal(err.Error())
		}
		os.Exit(2)
	}
}

var cmdAddonServices = &Command{
	Run:      runAddonServices,
	Usage:    "addon-services",
	Category: "add-on",
	Short:    "list addon services" + extra,
	Long: `
Lists available addon services.

Examples:

    $ hk addon-services
    heroku-postgresql
    newrelic
    redisgreen
    ...
`,
}

func runAddonServices(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	services, err := client.AddonServiceList(nil)
	must(err)

	for _, s := range services {
		fmt.Println(s.Name)
	}
}

var cmdAddonPlans = &Command{
	Run:      runAddonPlans,
	Usage:    "addon-plans <service>",
	Category: "add-on",
	Short:    "list addon plans" + extra,
	Long: `
Lists available addon plans for an addon provider.

Examples:

    $ hk addon-plans heroku-postgresql
    hobby-dev        $0/mo
    hobby-basic      $9/mo
    standard-yanari  $50/mo
    standard-tengu   $200/mo
    premium-yanari   $200/mo
    premium-tengu    $350/mo
    standard-ika     $750/mo
    premium-ika      $1200/mo
    ...
`,
}

func runAddonPlans(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	service := args[0]
	plans, err := client.PlanList(service, nil)
	must(err)

	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	sort.Sort(addonPlansByPrice(plans))
	for _, p := range plans {
		listRec(w,
			strings.TrimPrefix(p.Name, service+":"),
			addonPlanPriceString(p),
		)
	}
}

type addonPlansByPrice []heroku.Plan

func (a addonPlansByPrice) Len() int           { return len(a) }
func (a addonPlansByPrice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a addonPlansByPrice) Less(i, j int) bool { return a[i].Price.Cents < a[j].Price.Cents }

func addonPlanPriceString(p heroku.Plan) string {
	r := big.NewRat(int64(p.Price.Cents), 100)
	decimals := 2
	if p.Price.Cents%100 == 0 {
		decimals = 0
	}
	return "$" + r.FloatString(decimals) + "/" + shortenPriceUnit(p.Price.Unit)
}

func shortenPriceUnit(unit string) string {
	switch unit {
	case "month":
		return "mo"
	default:
		return unit
	}
}
