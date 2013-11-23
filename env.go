package main

import (
	"fmt"
	"log"
	"sort"
	"strings"
)

var nsEnv = &Namespace{
	Name: "env",
	Commands: []*Command{
		cmdEnvList,
		cmdEnvGet,
		cmdEnvSet,
		cmdEnvUnset,
	},
	Short: "manage app config",
}

var cmdEnvList = &Command{
	Run:   runEnvList,
	Usage: "list",
	Short: "list config vars",
	Long:  `Show all config vars.`,
}

func runEnvList(cmd *Command, args []string) {
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

var cmdEnvGet = &Command{
	Run:   runEnvGet,
	Usage: "get <name>",
	Short: "get config var" + extra,
	Long: `
Get the value of a config var.

Example:

  $ hk get BUILDPACK_URL
  http://github.com/kr/heroku-buildpack-inline.git
`,
}

func runEnvGet(cmd *Command, args []string) {
	if len(args) != 1 {
		log.Fatal("Invalid usage. See 'hk help env get'")
	}
	config, err := client.ConfigVarInfo(mustApp())
	must(err)
	value, found := config[args[0]]
	if !found {
		log.Fatalf("No such key as '%s'", args[0])
	}
	fmt.Println(value)
}

var cmdEnvSet = &Command{
	Run:   runEnvSet,
	Usage: "set <name>=<value> ...",
	Short: "set config var",
	Long: `
Set the value of a config var.

Example:

  $ hk set BUILDPACK_URL=http://github.com/kr/heroku-buildpack-inline.git
`,
}

func runEnvSet(cmd *Command, args []string) {
	if len(args) < 1 {
		log.Fatal("Invalid usage. See 'hk help env set'")
	}
	config := make(map[string]*string)
	for _, arg := range args {
		i := strings.Index(arg, "=")
		if i < 0 {
			log.Fatalf("bad format: %#q. See 'hk help env set'", arg)
		}
		val := arg[i+1:]
		config[arg[:i]] = &val
	}
	_, err := client.ConfigVarUpdate(mustApp(), config)
	must(err)
}

var cmdEnvUnset = &Command{
	Run:   runEnvUnset,
	Usage: "unset <name> ...",
	Short: "unset config var",
	Long: `
Unset a config var.

Example:

  $ hk unset BUILDPACK_URL
`,
}

func runEnvUnset(cmd *Command, args []string) {
	if len(args) < 1 {
		log.Fatal("Invalid usage. See 'hk help env unset'")
	}
	config := make(map[string]*string)
	for _, key := range args {
		config[key] = nil
	}
	_, err := client.ConfigVarUpdate(mustApp(), config)
	must(err)
}
