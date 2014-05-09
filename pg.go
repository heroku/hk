package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/bgentry/heroku-go"
	"github.com/heroku/hk/postgresql"
)

var cmdPgList = &Command{
	Run:      runPgList,
	Usage:    "pg-list",
	NeedsApp: true,
	Category: "pg",
	Short:    "list Heroku Postgres databases" + extra,
	Long: `
Pg-list shows the name, plan, state, and connection count for
all Heroku Postgres databases on an app. Forks and followers are
shown in a tree under the database they follow.

The database configured as your DATABASE_URL is indicated with
an asterisk (*). Exclamation marks (!!) indicate databases which
are due for maintenance.

Examples:

    $ hk pg-list
    * heroku-postgresql-crimson       crane  available  5
      └───> heroku-postgresql-copper  ronin  available  3

    $ hk pg-list
      heroku-postgresql-green              standard-tengu  available     3
    * heroku-postgresql-olive              standard-tengu  available     3
      ├───> heroku-postgresql-gray         standard-tengu  available !!  3
      ├─ ─┤ heroku-postgresql-rose         standard-tengu  available     3
      │     └───> heroku-postgresql-white  standard-tengu  available     3
      └─ ─┤ heroku-postgresql-teal         standard-tengu  available     3
`,
}

func runPgList(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	appname := mustApp()
	// list all addons
	addons, err := client.AddonList(appname, nil)
	must(err)

	// locate Heroku Postgres addons
	hpgprefix := hpgAddonName() + "-"
	hpgs := make(map[string]*heroku.Addon)
	for i := range addons {
		if strings.HasPrefix(addons[i].Name, hpgprefix) {
			hpgs[addons[i].Name] = &addons[i]
		}
	}
	if len(hpgs) == 0 {
		return // no Heroku Postgres databases to list
	}

	// fetch app's config concurrently in case we need to resolve DB names
	var appConf map[string]string
	confch := make(chan map[string]string, 1)
	errch := make(chan error, len(hpgs)+1)
	go func(appname string) {
		if config, err := client.ConfigVarInfo(appname); err != nil {
			errch <- err
		} else {
			confch <- config
		}
	}(appname)

	// fetch info for each database concurrently
	var dbinfos []*fullDBInfo
	dbinfoch := make(chan fullDBInfo, len(hpgs))
	for name, addon := range hpgs {
		go func(name string, addon *heroku.Addon) {
			db := pgclient.NewDB(addon.ProviderId, addon.Plan.Name)
			if dbinfo, err := db.Info(); err != nil {
				errch <- err
			} else {
				dbinfoch <- fullDBInfo{Name: name, DBInfo: dbinfo}
			}
		}(name, addon)
	}
	// wait for db info repsonses and app config response
	for i := 0; i < len(hpgs)+1; i++ {
		select {
		case err := <-errch:
			printFatal(err.Error())
		case dbinfo := <-dbinfoch:
			dbinfos = append(dbinfos, &dbinfo)
		case appConf = <-confch:
		}
	}

	addonMap := newPgAddonMap(addons, appConf)
	dbinfos = sortedDBInfoTree(dbinfos, addonMap)

	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()
	printDBTree(w, dbinfos, addonMap)
}

var cmdPgInfo = &Command{
	Run:      runPgInfo,
	Usage:    "pg-info <dbname>",
	NeedsApp: true,
	Category: "pg",
	Short:    "show Heroku Postgres database info" + extra,
	Long: `
Pg-info shows general information about a Heroku Postgres
database.

Examples:

    $ hk pg-info heroku-postgresql-crimson
    Name:         heroku-postgresql-crimson
    Env Vars:     DATABASE_URL, HEROKU_POSTGRESQL_CRIMSON_URL
    Plan:         Crane
    Status:       Available
    Data Size:    6.3 MB
    Tables:       3
    PG Version:   9.1.11
    Connections:  5
    Fork/Follow:  Available
    Rollback:     Unsupported
    Created:      2013-11-19 20:40 UTC
    Followers:    none
    Forks:        heroku-postgresql-copper
    Maintenance:  not required

    $ hk pg-info crimson
    ...
`,
}

func runPgInfo(cmd *Command, args []string) {
	if len(args) > 1 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	appname := mustApp()
	var addonName string
	if len(args) > 0 {
		addonName = ensurePrefix(args[0], hpgAddonName()+"-")
	}

	_, dbi, addonMap := mustGetDBInfoAndAddonMap(addonName, appname)
	printPgInfo(addonName, dbi, &addonMap)
}

func printPgInfo(name string, dbi postgresql.DBInfo, addonMap *pgAddonMap) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	listRec(w, "Name:", name)
	envNames := strings.Join(addonMap.FindEnvsFromValue(dbi.ResourceURL), ", ")
	listRec(w, "Env Vars:", envNames)

	// List info items returned by PG API
	for _, ie := range dbi.Info {
		if len(ie.Values) == 0 {
			listRec(w, ie.Name+":", "none")
		} else {
			for n, val := range ie.Values {
				label := ie.Name + ":"
				if n != 0 {
					label = ""
				}
				// try to resolve the value to an addon name if PG API says we should
				if ie.ResolveDBName {
					valstr := val.(string)
					if addonName, ok := addonMap.FindAddonFromValue(valstr); ok {
						// resolved it to an addon name, print that instead
						valstr = addonName
					} else {
						// Couldn't resolve to an addon name. Try to parse the URL so we
						// can display only its Host and Path (without creds).
						if u, err := url.Parse(valstr); err == nil && u.User != nil {
							valstr = u.Host + u.Path
						}
					}
					listRec(w, label, valstr)
					continue
				}
				listRec(w, label, val)
			}
		}
	}
}

var cmdPgUnfollow = &Command{
	Run:      runPgUnfollow,
	Usage:    "pg-unfollow <dbname>",
	NeedsApp: true,
	Category: "pg",
	Short:    "stop a replica postgres database from following" + extra,
	Long: `
Pg-unfollow stops a Heroku Postgres database follower from
following, turning it into a read/write database. The command
will prompt for confirmation, or accept confirmation via stdin.

Examples:

    $ hk pg-unfollow heroku-postgresql-blue
    warning: heroku-postgresql-blue on myapp will permanently stop following heroku-postgresql-red.
    warning: This cannot be undone. Please type "heroku-postgresql-blue" to continue:
    > heroku-postgresql-blue
    Unfollowed heroku-postgresql-blue on myapp.

    $ hk pg-unfollow blue
    warning: heroku-postgresql-blue on myapp will permanently stop following heroku-postgresql-red.
    warning: This cannot be undone. Please type "blue" to continue:
    > blue
    Unfollowed heroku-postgresql-blue on myapp.

    $ echo blue | hk pg-unfollow blue
    Unfollowed heroku-postgresql-blue on myapp.
`,
}

func runPgUnfollow(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.PrintUsage()
		os.Exit(2)
	}
	appname := mustApp()
	addonName := ensurePrefix(args[0], hpgAddonName()+"-")

	db, dbi, addonMap := mustGetDBInfoAndAddonMap(addonName, appname)
	if !dbi.IsFollower() {
		printFatal("%s is not following another database.", addonName)
	}
	parentName := getResolvedInfoValue(dbi, "Following", &addonMap)

	printWarning("%s on %s will permanently stop following %s.", addonName, appname, parentName)
	warning := fmt.Sprintf("This cannot be undone. Please type %q to continue:", args[0])
	mustConfirm(warning, args[0])

	must(db.Unfollow())
	fmt.Printf("Unfollowed %s on %s.\n", addonName, appname)
}

var commandNamePsql string

var cmdPsql = &Command{
	Run:      runPsql,
	Usage:    "psql [-c <command>] [<dbname>]",
	NeedsApp: true,
	Category: "pg",
	Short:    "open a psql shell to a Heroku Postgres database" + extra,
	Long: `
Psql opens a PostgreSQL shell to a Heroku Postgres database
using the locally-installed psql command.

Examples:

    $ hk psql
    psql (9.3.1, server 9.1.11)
    SSL connection (cipher: DHE-RSA-AES256-SHA, bits: 256)
    Type "help" for help.
    
    d1234abcdefghi=>

    $ hk psql crimson
    ...

    $ hk psql heroku-postgresql-crimson
    ...
`,
}

func init() {
	cmdPsql.Flag.StringVarP(&commandNamePsql, "command", "c", "", "SQL command to run")
}

func runPsql(cmd *Command, args []string) {
	if len(args) > 1 {
		cmd.PrintUsage()
		os.Exit(2)
	}

	configName := "DATABASE_URL"
	if len(args) == 1 {
		configName = dbNameToPgEnv(args[0])
	}
	appname := mustApp()

	// Make sure psql is installed
	if _, err := exec.LookPath("psql"); err != nil {
		printFatal("Local psql command not found. For help installing psql, see http://devcenter.heroku.com/articles/local-postgresql")
	}

	// fetch app's config to get the URL
	config, err := client.ConfigVarInfo(appname)
	must(err)

	// get URL
	urlstr, exists := config[configName]
	if !exists {
		printFatal("Env %s not found", configName)
	}
	u, err := url.Parse(urlstr)
	if err != nil {
		printFatal("Invalid URL at env " + configName)
	}

	// handle custom port
	hostname := u.Host
	portnum := 5432
	if colIndex := strings.Index(u.Host, ":"); colIndex != -1 {
		hostname = u.Host[:colIndex]
		portnum, err = strconv.Atoi(u.Host[colIndex+1:])
		if err != nil {
			printFatal("Invalid port in %s: %s", configName, u.Host[colIndex+1:])
		}
	}

	if u.User == nil || u.User.Username() == "" {
		printFatal("Missing credentials in %s", configName)
	}

	// construct and run psql command
	psqlArgs := []string{
		"psql",
		"-U", u.User.Username(),
		"-h", hostname,
		"-p", strconv.Itoa(portnum),
	}
	if commandNamePsql != "" {
		psqlArgs = append(psqlArgs, "-c")
		psqlArgs = append(psqlArgs, commandNamePsql)
	}
	psqlArgs = append(psqlArgs, u.Path[1:])

	pgenv := os.Environ()
	pass, _ := u.User.Password()
	pgenv = append(pgenv, "PGPASSWORD="+pass)
	pgenv = append(pgenv, "PGSSLMODE=require")

	if err := runCommand("psql", psqlArgs, pgenv); err != nil {
		printFatal("Error running psql: %s", err)
	}
}
