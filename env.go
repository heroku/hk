package main

import (
	"fmt"
	"log"
	"sort"
	"strings"
)

var cmdEnv = &Command{
	Run:      runEnv,
	Name:     "env",
	Usage:    "[-a <app>]",
	Category: "config",
	Short:    "list env vars",
	Long:     `Show all env vars.`,
}

func init() {
	cmdEnv.Flag.StringVar(&flagApp, "a", "", "app name")
}

func runEnv(cmd *Command, args []string) {
	config, err := client.ConfigVarInfo(mustApp())
	must(err)
	var configKeys []string
	for k := range config {
		configKeys = append(configKeys, k)
	}
	sort.Strings(configKeys)
	for _, k := range configKeys {
		fmt.Printf("%s=%s\n", k, config[k])
	}
}

var cmdGet = &Command{
	Run:      runGet,
	Name:     "get",
	Usage:    "[-a <app>] <name>",
	Category: "config",
	Short:    "get env var" + extra,
	Long: `
Get the value of an env var.

Example:

  $ hk get BUILDPACK_URL
  http://github.com/kr/heroku-buildpack-inline.git
`,
}

func init() {
	cmdGet.Flag.StringVar(&flagApp, "a", "", "app name")
}

func runGet(cmd *Command, args []string) {
	if len(args) != 1 {
		log.Fatal("Invalid usage. See 'hk help get'")
	}
	config, err := client.ConfigVarInfo(mustApp())
	must(err)
	value, found := config[args[0]]
	if !found {
		log.Fatalf("No such key as '%s'", args[0])
	}
	fmt.Println(value)
}

var cmdSet = &Command{
	Run:      runSet,
	Name:     "set",
	Usage:    "[-a <app>] <name>=<value>...",
	Category: "config",
	Short:    "set env var",
	Long: `
Set the value of a env var.

Example:

  $ hk set BUILDPACK_URL=http://github.com/kr/heroku-buildpack-inline.git
`,
}

func init() {
	cmdSet.Flag.StringVar(&flagApp, "a", "", "app name")
}

func runSet(cmd *Command, args []string) {
	if len(args) < 1 {
		log.Fatal("Invalid usage. See 'hk help set'")
	}
	config := make(map[string]*string)
	for _, arg := range args {
		i := strings.Index(arg, "=")
		if i < 0 {
			log.Fatalf("bad format: %#q. See 'hk help set'", arg)
		}
		val := arg[i+1:]
		config[arg[:i]] = &val
	}
	_, err := client.ConfigVarUpdate(mustApp(), config)
	must(err)
}

var cmdUnset = &Command{
	Run:      runUnset,
	Name:     "unset",
	Usage:    "[-a <app>] <name>...",
	Category: "config",
	Short:    "unset env var",
	Long: `
Unset a env var.

Example:

  $ hk unset BUILDPACK_URL
`,
}

func init() {
	cmdUnset.Flag.StringVar(&flagApp, "a", "", "app name")
}

func runUnset(cmd *Command, args []string) {
	if len(args) < 1 {
		log.Fatal("Invalid usage. See 'hk help unset'")
	}
	config := make(map[string]*string)
	for _, key := range args {
		config[key] = nil
	}
	_, err := client.ConfigVarUpdate(mustApp(), config)
	must(err)
}
