package main

import (
	"fmt"
	"log"
	"sort"
	"strings"
)

var cmdEnv = &Command{
	Run:   runEnv,
	Usage: "env [-a app]",
	Short: "list config vars",
	Long:  `Show all config vars.`,
}

func init() {
	cmds := []*Command{cmdEnv, cmdGet, cmdSet}
	for _, c := range cmds {
		c.Flag.StringVar(&flagApp, "a", "", "app")
	}
}

func runEnv(cmd *Command, args []string) {
	var config map[string]string
	APIReq("GET", "/apps/"+mustApp()+"/config_vars").Do(&config)
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
	Run:   runGet,
	Usage: "get [-a app] <name>",
	Short: "get config var",
	Long: `
Get the value of a config var.

Example:

  $ hk get BUILDPACK_URL
  http://github.com/kr/heroku-buildpack-inline.git
`,
}

func runGet(cmd *Command, args []string) {
	if len(args) != 1 {
		log.Fatal("Invalid usage. See 'hk help get'")
	}
	var config map[string]string
	APIReq("GET", "/apps/"+mustApp()+"/config_vars").Do(&config)
	value, found := config[args[0]]
	if !found {
		log.Fatalf("No such key as '%s'", args[0])
	}
	fmt.Println(value)
}

var cmdSet = &Command{
	Run:   runSet,
	Usage: "set [-a app] <name>=<value> ...",
	Short: "set config var",
	Long: `
Set the value of a config var.

Example:

  $ hk set BUILDPACK_URL=http://github.com/kr/heroku-buildpack-inline.git
`,
}

func runSet(cmd *Command, args []string) {
	config := make(map[string]string)
	for _, arg := range args {
		i := strings.Index(arg, "=")
		if i < 0 {
			log.Fatalf("bad format: %#q. See 'hk help set'", arg)
		}
		config[arg[:i]] = arg[i+1:]
	}
	r := APIReq("PUT", "/apps/"+mustApp()+"/config_vars")
	r.SetBodyJson(config)
	r.Do(nil)
}
