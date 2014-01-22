package main

import (
	"net/url"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/heroku/hk/postgresql"
)

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
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	appname := mustApp()
	addonName := ensurePrefix(args[0], "heroku-postgresql-")
	addon, err := client.AddonInfo(appname, addonName)
	must(err)

	// fetch app's config concurrently in case we need to resolve DB names
	var appConf map[string]string
	confch := make(chan map[string]string, 1)
	errch := make(chan error, 1)
	go func(appname string) {
		if config, err := client.ConfigVarInfo(appname); err != nil {
			errch <- err
		} else {
			confch <- config
		}
	}(appname)

	db := pgclient.NewDB(addon.ProviderId, addon.Plan.Name)
	info, err := db.Info()
	must(err)

	select {
	case err := <-errch:
		printError(err.Error())
	case appConf = <-confch:
	}

	printPgInfo(addonName, info, appConf)
}

func printPgInfo(name string, info postgresql.DBInfo, appConf map[string]string) {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()

	listRec(w, "Name:", name)
	envNames := strings.Join(EnvNamesFromURL(info.ResourceURL, appConf), ", ")
	listRec(w, "Env Vars:", envNames)

	// List info items returned by PG API
	for _, ie := range info.Info {
		if len(ie.Values) == 0 {
			listRec(w, ie.Name+":", "none")
		} else {
			for n, val := range ie.Values {
				label := ie.Name + ":"
				if n != 0 {
					label = ""
				}
				if ie.ResolveDBName {
					valstr := val.(string)
					listRec(w, label, DatabaseNameFromURL(valstr, appConf))
				} else {
					listRec(w, label, val)
				}
			}
		}
	}
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
	cmdPsql.Flag.StringVar(&commandNamePsql, "c", "", "SQL command to run")
}

func runPsql(cmd *Command, args []string) {
	if len(args) > 1 {
		cmd.printUsage()
		os.Exit(2)
	}

	configName := "DATABASE_URL"
	if len(args) == 1 {
		configName = dbNameToPgEnv(args[0])
	}
	appname := mustApp()

	// Make sure psql is installed
	if _, err := exec.LookPath("psql"); err != nil {
		printError("Local psql command not found. For help installing psql, see http://devcenter.heroku.com/articles/local-postgresql")
	}

	// fetch app's config to get the URL
	config, err := client.ConfigVarInfo(appname)
	must(err)

	// get URL
	urlstr, exists := config[configName]
	if !exists {
		printError("Env %s not found", configName)
	}
	u, err := url.Parse(urlstr)
	if err != nil {
		printError("Invalid URL at env " + configName)
	}

	// handle custom port
	hostname := u.Host
	portnum := 5432
	if colIndex := strings.Index(u.Host, ":"); colIndex != -1 {
		hostname = u.Host[:colIndex]
		portnum, err = strconv.Atoi(u.Host[colIndex+1:])
		if err != nil {
			printError("Invalid port in %s: %s", configName, u.Host[colIndex+1:])
		}
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
		printError("Error running psql: %s", err)
	}
}

func EnvNamesFromURL(url string, env map[string]string) (names []string) {
	for k, v := range env {
		if isEnvKeyFromPostgres(k) && v == url {
			names = append(names, k)
		}
	}
	sort.Strings(names)
	return
}

func DatabaseNameFromURL(url string, env map[string]string) string {
	for k, v := range env {
		if isEnvKeyFromPostgres(k) && v == url {
			return pgEnvToDBName(k)
		}
	}
	return url
}

func pgEnvToDBName(key string) string {
	return strings.ToLower(strings.Replace(strings.TrimSuffix(key, "_URL"), "_", "-", -1))
}

func dbNameToPgEnv(name string) string {
	return ensurePrefix(
		strings.ToUpper(strings.Replace(name, "-", "_", -1)),
		"HEROKU_POSTGRESQL_",
	) + "_URL"
}

func isEnvKeyFromPostgres(key string) bool {
	return key == "DATABASE_URL" ||
		strings.HasPrefix(key, "HEROKU_POSTGRESQL_") && strings.HasSuffix(key, "_URL")
}
